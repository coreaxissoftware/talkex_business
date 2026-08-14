// Package messenger implements the Facebook Messenger connector via the
// Send API. Currently simulated; production will use the official
// Messenger Platform Send API with page access tokens.
package messenger

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

func (c *Connector) Name() string { return "messenger" }

func (c *Connector) ValidateConfig(cfg map[string]string) error {
	if cfg["page_access_token"] == "" {
		return fmt.Errorf("messenger: page_access_token is required")
	}
	return nil
}

func (c *Connector) Send(msg *shared.OutboundMessage) (*shared.DeliveryResult, error) {
	// Simulated: in production this calls the Messenger Send API
	time.Sleep(time.Duration(50+rand.Intn(100)) * time.Millisecond)

	if rand.Intn(15) == 0 {
		return nil, fmt.Errorf("messenger: simulated delivery failure")
	}

	externalID := fmt.Sprintf("fb_%d_%s", time.Now().UnixNano(), msg.ID[:8])
	log.Printf("messenger: [SIM] message %s → %s (body: %.50s...)",
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
