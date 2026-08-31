package rcs

// Rich message payloads for RCS Business Messaging (parity with Gupshup's
// RCS agent). These are pure serialisation types — the connector's Send
// method already POSTs the outbound payload; the RichCard / RichCarousel
// helpers let a template or campaign describe the payload declaratively
// and get JSON that RCS platforms accept in Google Jibe & most Indian
// telcos' RCS bureaus.

// RichCard is one card in a rich message. Height determines the card
// preview aspect ratio: SHORT (112px), MEDIUM (168px), TALL (264px).
type RichCard struct {
	Title           string          `json:"title"`
	Description     string          `json:"description,omitempty"`
	MediaURL        string          `json:"media_url,omitempty"`
	MediaHeight     string          `json:"media_height,omitempty"` // SHORT | MEDIUM | TALL
	SuggestedReplies []SuggestedReply `json:"suggested_replies,omitempty"`
	SuggestedActions []SuggestedAction `json:"suggested_actions,omitempty"`
}

// SuggestedReply — one-tap reply chip. Text becomes the inbound
// message body when tapped; PostbackData travels back invisibly for
// analytics.
type SuggestedReply struct {
	Text         string `json:"text"`
	PostbackData string `json:"postback_data,omitempty"`
}

// SuggestedAction — one-tap link that opens a URL, dials a number, or
// starts a calendar invite.
type SuggestedAction struct {
	Type         string `json:"type"` // openUrl | dialPhone | createCalendar
	Text         string `json:"text"`
	URL          string `json:"url,omitempty"`
	Phone        string `json:"phone,omitempty"`
	PostbackData string `json:"postback_data,omitempty"`
}

// RichCarousel — up to 10 cards displayed as a horizontal-swipe stack.
// CardWidth controls the pixel width (SMALL 136px, MEDIUM 232px).
type RichCarousel struct {
	CardWidth string     `json:"card_width"`
	Cards     []RichCard `json:"cards"`
}

// BuildStandaloneCardPayload returns the JSON body the Send() method
// should POST when the message is a single rich card. Kept as a helper
// so the outbound queue can produce this shape without importing the
// upstream provider SDK.
func BuildStandaloneCardPayload(card RichCard) map[string]interface{} {
	if card.MediaHeight == "" {
		card.MediaHeight = "MEDIUM"
	}
	return map[string]interface{}{
		"contentMessage": map[string]interface{}{
			"richCard": map[string]interface{}{
				"standaloneCard": map[string]interface{}{
					"cardOrientation": "VERTICAL",
					"cardContent":     card,
				},
			},
		},
	}
}

// BuildCarouselPayload returns the JSON body for a multi-card carousel.
func BuildCarouselPayload(c RichCarousel) map[string]interface{} {
	if c.CardWidth == "" {
		c.CardWidth = "MEDIUM"
	}
	return map[string]interface{}{
		"contentMessage": map[string]interface{}{
			"richCard": map[string]interface{}{
				"carouselCard": c,
			},
		},
	}
}
