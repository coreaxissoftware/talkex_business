// Package ai wraps the Anthropic Claude SDK for in-app AI assist features:
// reply suggestions, conversation summaries, and sentiment tagging.
//
// If ANTHROPIC_API_KEY is unset the service returns simulated responses
// so the UI works in dev without a paid key — matches the OTP + Razorpay
// dev-mode pattern used elsewhere in this codebase.
package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Shared client — the SDK is designed to be constructed once and
// reused across requests, so we build lazily on the first real call.
var (
	sharedClient     *anthropic.Client
	sharedClientOnce sync.Once
)

func client() *anthropic.Client {
	sharedClientOnce.Do(func() {
		c := anthropic.NewClient(option.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")))
		sharedClient = &c
	})
	return sharedClient
}

// callTimeout is the maximum wall-clock a single AI request may run.
// Wraps the caller's ctx so a hung upstream cannot tie up a goroutine
// (and paid inference minutes) for the SDK's 10-minute default.
const callTimeout = 30 * time.Second

// Model is the Claude model used for every AI call. Opus 5 is the
// default per the claude-api skill guidance; users can override via
// ANTHROPIC_MODEL for cost tuning.
func modelID() anthropic.Model {
	if m := os.Getenv("ANTHROPIC_MODEL"); m != "" {
		return anthropic.Model(m)
	}
	return anthropic.Model("claude-opus-5")
}

// devMode returns true when no API key is configured.
func devMode() bool {
	return os.Getenv("ANTHROPIC_API_KEY") == ""
}

// Message is a normalized turn in a conversation — direction plus body.
type Message struct {
	Direction string    `json:"direction"` // inbound | outbound
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

// SuggestReply asks Claude for a professional next reply given the
// last N messages of a conversation.
func SuggestReply(ctx context.Context, history []Message, contactName string) (string, error) {
	if devMode() {
		return devSuggestion(history, contactName), nil
	}

	transcript := renderTranscript(history)
	prompt := fmt.Sprintf(`You are a support agent replying on behalf of a business on WhatsApp/Chat.
The contact you're replying to is: %s
Below is the conversation so far. Suggest ONE concise, professional next reply from the AGENT (outbound).
Keep it under 40 words. No preamble, no quotation marks — just the reply text.

Conversation:
%s`, contactName, transcript)

	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	resp, err := client().Messages.New(callCtx, anthropic.MessageNewParams{
		Model:     modelID(),
		MaxTokens: 512,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", err
	}
	return extractText(resp), nil
}

// Summarize returns a 2-3 sentence summary of a conversation for
// handoff between agents.
func Summarize(ctx context.Context, history []Message, contactName string) (string, error) {
	if devMode() {
		return devSummary(history, contactName), nil
	}

	transcript := renderTranscript(history)
	prompt := fmt.Sprintf(`Summarize this customer conversation in 2-3 short sentences.
Focus on: what the customer wants, what has been resolved, what's still pending.
No preamble. Contact: %s

Conversation:
%s`, contactName, transcript)

	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	resp, err := client().Messages.New(callCtx, anthropic.MessageNewParams{
		Model:     modelID(),
		MaxTokens: 512,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", err
	}
	return extractText(resp), nil
}

// Sentiment returns "positive", "neutral", or "negative" plus a
// one-line explanation.
type SentimentResult struct {
	Score  string `json:"score"`
	Reason string `json:"reason"`
}

func Sentiment(ctx context.Context, history []Message) (*SentimentResult, error) {
	if devMode() {
		return devSentiment(history), nil
	}

	transcript := renderTranscript(history)
	prompt := fmt.Sprintf(`Classify the customer's sentiment in this conversation.
Respond with EXACTLY two lines:
Line 1: one word — positive, neutral, or negative
Line 2: a short (max 20 words) reason

Conversation:
%s`, transcript)

	callCtx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	resp, err := client().Messages.New(callCtx, anthropic.MessageNewParams{
		Model:     modelID(),
		MaxTokens: 128,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return nil, err
	}
	text := strings.TrimSpace(extractText(resp))
	lines := strings.SplitN(text, "\n", 2)
	res := &SentimentResult{}
	if len(lines) > 0 {
		res.Score = strings.ToLower(strings.TrimSpace(lines[0]))
	}
	if len(lines) > 1 {
		res.Reason = strings.TrimSpace(lines[1])
	}
	return res, nil
}

// renderTranscript formats history for the model prompt.
func renderTranscript(msgs []Message) string {
	var b strings.Builder
	for _, m := range msgs {
		who := "AGENT"
		if m.Direction == "inbound" {
			who = "CUSTOMER"
		}
		b.WriteString(fmt.Sprintf("%s: %s\n", who, m.Body))
	}
	return b.String()
}

// extractText pulls concatenated text from a response.
func extractText(resp *anthropic.Message) string {
	var b strings.Builder
	for _, block := range resp.Content {
		if t, ok := block.AsAny().(anthropic.TextBlock); ok {
			b.WriteString(t.Text)
		}
	}
	return strings.TrimSpace(b.String())
}

// ---- Dev-mode simulations -------------------------------------------

func devSuggestion(history []Message, contactName string) string {
	last := ""
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Direction == "inbound" {
			last = strings.ToLower(history[i].Body)
			break
		}
	}
	first := strings.SplitN(contactName, " ", 2)[0]
	if first == "" {
		first = "there"
	}
	switch {
	case strings.Contains(last, "price") || strings.Contains(last, "cost"):
		return fmt.Sprintf("Hi %s, thanks for asking! Could you share your requirements so I can send you an accurate quote?", first)
	case strings.Contains(last, "when") || strings.Contains(last, "delivery"):
		return fmt.Sprintf("Hi %s, delivery typically takes 2-3 business days from confirmation. Would you like me to check your order status?", first)
	case strings.Contains(last, "thank"):
		return fmt.Sprintf("You're most welcome, %s! Let us know if there's anything else we can help with.", first)
	case strings.Contains(last, "hi") || strings.Contains(last, "hello"):
		return fmt.Sprintf("Hi %s! How can we help you today?", first)
	default:
		return fmt.Sprintf("Thanks for reaching out, %s. Could you tell me a bit more about what you're looking for?", first)
	}
}

func devSummary(history []Message, contactName string) string {
	inbound, outbound := 0, 0
	for _, m := range history {
		if m.Direction == "inbound" {
			inbound++
		} else {
			outbound++
		}
	}
	return fmt.Sprintf("Conversation with %s across %d messages (%d from customer, %d agent replies). Configure ANTHROPIC_API_KEY for real AI summaries.",
		contactName, inbound+outbound, inbound, outbound)
}

func devSentiment(history []Message) *SentimentResult {
	positive := []string{"thanks", "great", "love", "awesome", "perfect", "excellent"}
	negative := []string{"angry", "bad", "terrible", "hate", "worst", "refund", "cancel", "poor"}
	pos, neg := 0, 0
	for _, m := range history {
		if m.Direction != "inbound" {
			continue
		}
		body := strings.ToLower(m.Body)
		for _, w := range positive {
			if strings.Contains(body, w) {
				pos++
			}
		}
		for _, w := range negative {
			if strings.Contains(body, w) {
				neg++
			}
		}
	}
	if pos > neg {
		return &SentimentResult{Score: "positive", Reason: "Customer expressed satisfaction (dev heuristic)."}
	}
	if neg > pos {
		return &SentimentResult{Score: "negative", Reason: "Customer expressed frustration or dissatisfaction (dev heuristic)."}
	}
	return &SentimentResult{Score: "neutral", Reason: "No strong sentiment cues detected (dev heuristic)."}
}
