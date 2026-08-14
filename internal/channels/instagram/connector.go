// Package instagram implements the Instagram DM connector via the
// Instagram Graph API / Messenger Platform. Currently simulated;
// production will use the official Instagram Messaging API.
package instagram

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/channels/shared"
)

type Connector struct {
	PageAccessToken string
}

func (c *Connector) Name() string { return "instagram" }

func (c *Connector) ValidateConfig(cfg map[string]string) error {
	if cfg["page_access_token"] == "" {
		return fmt.Errorf("instagram: page_access_token is required")
	}
	return nil
}

func (c *Connector) Send(msg *shared.OutboundMessage) (*shared.DeliveryResult, error) {
	// Simulated: in production this calls the Instagram Messaging API
	time.Sleep(time.Duration(60+rand.Intn(140)) * time.Millisecond)

	if rand.Intn(15) == 0 {
		return nil, fmt.Errorf("instagram: simulated delivery failure")
	}

	externalID := fmt.Sprintf("ig_%d_%s", time.Now().UnixNano(), msg.ID[:8])
	log.Printf("instagram: [SIM] message %s → %s (body: %.50s...)",
		msg.ID, msg.ContactID, msg.Body)

	return &shared.DeliveryResult{
		ExternalID: externalID,
		Status:     shared.StatusSent,
		Timestamp:  time.Now(),
	}, nil
}

func init() {
	shared.Register(&Connector{})
}
