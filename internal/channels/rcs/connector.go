// Package rcs implements the RCS Business Messaging connector.
// This is currently a simulated connector; production will integrate
// with Google's RCS Business Messaging API or a local aggregator.
package rcs

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/channels/shared"
)

type Connector struct {
	AgentID string
}

func (c *Connector) Name() string { return "rcs" }

func (c *Connector) ValidateConfig(cfg map[string]string) error {
	if cfg["agent_id"] == "" {
		return fmt.Errorf("rcs: agent_id is required")
	}
	return nil
}

func (c *Connector) Send(msg *shared.OutboundMessage) (*shared.DeliveryResult, error) {
	// Simulated: in production this calls the RCS Business Messaging API
	time.Sleep(time.Duration(80+rand.Intn(120)) * time.Millisecond)

	if rand.Intn(20) == 0 {
		return nil, fmt.Errorf("rcs: simulated delivery failure")
	}

	externalID := fmt.Sprintf("rcs_%d_%s", time.Now().UnixNano(), msg.ID[:8])
	log.Printf("rcs: [SIM] message %s → %s (body: %.50s...)",
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
