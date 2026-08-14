// Package email implements an SMTP email connector for transactional
// and campaign email delivery.
package email

import (
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/channels/shared"
)

// Connector sends messages via SMTP.
type Connector struct {
	Host     string // SMTP host e.g. smtp.gmail.com
	Port     string // SMTP port e.g. 587
	Username string
	Password string
	From     string // From address
}

func (c *Connector) Name() string { return "email" }

func (c *Connector) ValidateConfig(cfg map[string]string) error {
	required := []string{"smtp_host", "smtp_port", "smtp_user", "smtp_pass", "from_address"}
	for _, k := range required {
		if cfg[k] == "" {
			return fmt.Errorf("email: %s is required", k)
		}
	}
	return nil
}

func (c *Connector) Send(msg *shared.OutboundMessage) (*shared.DeliveryResult, error) {
	if c.Host == "" || c.From == "" {
		return nil, fmt.Errorf("email: SMTP not configured")
	}

	to := msg.ContactID // contact ID is the email address for email channel

	subject := "Message from TalkEx"
	if msg.Type == shared.TypeTemplate {
		subject = "Notification from TalkEx"
	}

	// Build MIME message
	var body strings.Builder
	body.WriteString(fmt.Sprintf("From: %s\r\n", c.From))
	body.WriteString(fmt.Sprintf("To: %s\r\n", to))
	body.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	body.WriteString("MIME-Version: 1.0\r\n")
	body.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	body.WriteString("\r\n")
	body.WriteString(msg.Body)

	auth := smtp.PlainAuth("", c.Username, c.Password, c.Host)
	addr := fmt.Sprintf("%s:%s", c.Host, c.Port)

	if err := smtp.SendMail(addr, auth, c.From, []string{to}, []byte(body.String())); err != nil {
		return nil, fmt.Errorf("email: smtp send failed: %w", err)
	}

	externalID := fmt.Sprintf("email_%d_%s", time.Now().UnixNano(), msg.ID[:8])

	return &shared.DeliveryResult{
		ExternalID: externalID,
		Status:     shared.StatusSent,
		Timestamp:  time.Now(),
	}, nil
}

func init() {
	shared.Register(&Connector{})
}
