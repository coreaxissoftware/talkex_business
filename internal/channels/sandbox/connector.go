// Package sandbox provides a simulated channel connector for developer
// testing. When sandbox mode is enabled in user settings, all outbound
// messages route through this connector instead of real channel APIs.
// Messages are logged but never actually delivered.
package sandbox

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/channels/shared"
)

type sandboxConnector struct{}

func (s *sandboxConnector) Name() string { return "sandbox" }

func (s *sandboxConnector) Send(msg *shared.OutboundMessage) (*shared.DeliveryResult, error) {
	// Simulate a small delay like a real API call
	time.Sleep(time.Duration(50+rand.Intn(150)) * time.Millisecond)

	// Simulate occasional failures for testing (10% failure rate)
	if rand.Intn(10) == 0 {
		return nil, fmt.Errorf("sandbox: simulated delivery failure for testing")
	}

	externalID := fmt.Sprintf("sandbox_%d_%s", time.Now().UnixNano(), msg.ID[:8])
	log.Printf("sandbox: [TEST] message %s → contact %s via %s (body: %.50s...)",
		msg.ID, msg.ContactID, msg.Channel, msg.Body)

	return &shared.DeliveryResult{
		Status:     shared.StatusSent,
		ExternalID: externalID,
	}, nil
}

func (s *sandboxConnector) ValidateConfig(cfg map[string]string) error {
	return nil // sandbox needs no config
}

func init() {
	shared.Register(&sandboxConnector{})
}
