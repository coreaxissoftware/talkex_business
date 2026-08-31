package flows

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

// postWebhook fires a small JSON POST to the given URL and returns true
// on any 2xx response. Used by the "webhook" journey step so a flow can
// call out to Zapier, Make, n8n, or an internal service at runtime.
func postWebhook(url, ownerID, contactID, stepID string, timeout time.Duration) bool {
	if url == "" {
		return false
	}
	body, _ := json.Marshal(map[string]string{
		"owner_id":   ownerID,
		"contact_id": contactID,
		"step_id":    stepID,
		"source":     "flow_webhook",
	})
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "TalkExFlowRunner/1.0")

	client := &http.Client{Timeout: timeout}
	res, err := client.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	return res.StatusCode >= 200 && res.StatusCode < 300
}
