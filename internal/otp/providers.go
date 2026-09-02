package otp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/config"
)

// Delivery providers for OTP codes. In dev mode (no credentials
// configured) the code is only logged; in production Mailgun handles
// email and MSG91 (India-first) or Twilio (global) handles SMS. Which
// SMS provider fires depends on which env vars are set — MSG91 wins
// when both are present because it's cheaper on Indian routes.

// deliverySender is the interface every real provider must satisfy.
// Kept private so we can swap implementations without leaking types.
type deliverySender interface {
	SendEmail(to, code string) error
	SendSMS(to, code string) error
}

// logOnly is the dev fallback — never returns an error, never delivers.
type logOnly struct{}

func (logOnly) SendEmail(to, code string) error {
	log.Printf("OTP (dev, email): %s → %s", to, code)
	return nil
}
func (logOnly) SendSMS(to, code string) error {
	log.Printf("OTP (dev, sms): %s → %s", to, code)
	return nil
}

// pickSender picks the right delivery backend based on which env
// vars are set. Deliberately re-runs per call so a runtime secret
// refresh works without a process restart.
func pickSender() deliverySender {
	cfg := config.Get()
	return combinedSender{
		email: pickEmail(cfg),
		sms:   pickSMS(cfg),
	}
}

// combinedSender routes email + SMS to potentially different backends.
// This lets a tenant use Mailgun for email but skip SMS in dev (the
// SMS side falls back to log-only when unconfigured).
type combinedSender struct {
	email deliverySender
	sms   deliverySender
}

func (c combinedSender) SendEmail(to, code string) error { return c.email.SendEmail(to, code) }
func (c combinedSender) SendSMS(to, code string) error   { return c.sms.SendSMS(to, code) }

// --- Email: Mailgun -------------------------------------------------

type mailgun struct {
	domain string
	apiKey string
	from   string
}

func (m mailgun) SendEmail(to, code string) error {
	if to == "" {
		return nil
	}
	form := url.Values{}
	form.Set("from", m.from)
	form.Set("to", to)
	form.Set("subject", "Your TalkEx verification code")
	form.Set("text",
		"Your TalkEx code is "+code+"\n\n"+
			"It expires in 10 minutes. If you didn't request this code, ignore this email.\n\n"+
			"— TalkEx Business")
	form.Set("html", fmt.Sprintf(
		`<div style="font-family:system-ui,-apple-system,sans-serif;max-width:480px;margin:auto;padding:32px 24px;background:#f7f3ee;border-radius:12px">`+
			`<h1 style="font-family:Georgia,serif;font-style:italic;color:#0f172a;margin:0 0 8px">TalkEx</h1>`+
			`<p style="color:#334155;margin:0 0 24px">Your verification code:</p>`+
			`<div style="font-family:'JetBrains Mono',ui-monospace,monospace;font-size:36px;letter-spacing:8px;background:#fff;padding:20px;border-radius:8px;text-align:center;color:#0ea5a0">%s</div>`+
			`<p style="color:#64748b;font-size:13px;margin-top:24px">Expires in 10 minutes. If you didn't request this, ignore this email.</p>`+
			`</div>`, code))

	// Mailgun US region URL. EU tenants set MAILGUN_BASE_URL to
	// https://api.eu.mailgun.net/v3/<domain>/messages.
	base := config.Get().MailgunBaseURL
	if base == "" {
		base = "https://api.mailgun.net/v3/" + m.domain + "/messages"
	}
	req, err := http.NewRequest(http.MethodPost, base, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth("api", m.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 8 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return fmt.Errorf("mailgun: HTTP %d", res.StatusCode)
	}
	return nil
}

// mailgun.SendSMS never runs — combinedSender routes SMS to the SMS
// sender — but the interface demands it.
func (mailgun) SendSMS(_, _ string) error { return nil }

func pickEmail(cfg *config.Config) deliverySender {
	if cfg.MailgunDomain != "" && cfg.MailgunAPIKey != "" {
		from := cfg.MailgunFrom
		if from == "" {
			from = "TalkEx <no-reply@" + cfg.MailgunDomain + ">"
		}
		return mailgun{domain: cfg.MailgunDomain, apiKey: cfg.MailgunAPIKey, from: from}
	}
	return logOnly{}
}

// --- SMS: MSG91 (India-first) or Twilio (global) --------------------

type msg91 struct {
	authKey    string
	templateID string
	senderID   string
	route      string
}

func (m msg91) SendSMS(to, code string) error {
	if to == "" {
		return nil
	}
	// MSG91 Flow API — the modern endpoint that respects DLT templates.
	// See https://docs.msg91.com/reference/sendotp — payload shape:
	//   { template_id, mobiles, VAR1: code }
	body, _ := json.Marshal(map[string]interface{}{
		"template_id": m.templateID,
		"mobiles":     normalisePhone(to),
		"otp":         code,
		"VAR1":        code,
		"sender":      m.senderID,
	})
	req, err := http.NewRequest(http.MethodPost, "https://control.msg91.com/api/v5/flow/", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("authkey", m.authKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 8 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return fmt.Errorf("msg91: HTTP %d", res.StatusCode)
	}
	return nil
}

func (msg91) SendEmail(_, _ string) error { return nil }

type twilioSMS struct {
	accountSID string
	authToken  string
	fromNumber string
}

func (t twilioSMS) SendSMS(to, code string) error {
	if to == "" {
		return nil
	}
	form := url.Values{}
	form.Set("From", t.fromNumber)
	form.Set("To", normalisePhone(to))
	form.Set("Body", "Your TalkEx verification code is "+code+". Valid for 10 minutes.")

	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.accountSID)
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(t.accountSID, t.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 8 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return fmt.Errorf("twilio: HTTP %d", res.StatusCode)
	}
	return nil
}

func (twilioSMS) SendEmail(_, _ string) error { return nil }

func pickSMS(cfg *config.Config) deliverySender {
	// MSG91 wins when configured — cheaper on Indian routes, DLT-native.
	if cfg.Msg91AuthKey != "" && cfg.Msg91TemplateID != "" {
		route := cfg.Msg91Route
		if route == "" {
			route = "4" // transactional
		}
		return msg91{
			authKey:    cfg.Msg91AuthKey,
			templateID: cfg.Msg91TemplateID,
			senderID:   cfg.Msg91SenderID,
			route:      route,
		}
	}
	if cfg.TwilioAccountSID != "" && cfg.TwilioAuthToken != "" && cfg.TwilioFromNumber != "" {
		return twilioSMS{
			accountSID: cfg.TwilioAccountSID,
			authToken:  cfg.TwilioAuthToken,
			fromNumber: cfg.TwilioFromNumber,
		}
	}
	return logOnly{}
}

// normalisePhone accepts either "9876543210" (assume +91) or "+919876543210"
// or "919876543210" and returns the E.164-with-plus form MSG91/Twilio
// expect. Handlers already trust the input's country prefix so this
// is a safety net, not a validator.
func normalisePhone(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return p
	}
	// Strip everything that isn't a digit or leading +
	var out strings.Builder
	for i, r := range p {
		if r == '+' && i == 0 {
			out.WriteRune(r)
			continue
		}
		if r >= '0' && r <= '9' {
			out.WriteRune(r)
		}
	}
	s := out.String()
	if strings.HasPrefix(s, "+") {
		return s
	}
	if len(s) == 10 {
		// Bare 10-digit → assume Indian mobile.
		return "+91" + s
	}
	// Already country-prefixed digits (e.g. 919876543210).
	return "+" + s
}
