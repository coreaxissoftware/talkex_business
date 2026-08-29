package metrics

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsHandlerFormat(t *testing.T) {
	// Reset counters so this test isn't order-dependent.
	countersMu.Lock()
	counters = map[string]*counter{}
	countersMu.Unlock()

	Describe("test_counter", "unit test")
	Inc("test_counter")
	Add("test_counter", 4)

	rec := httptest.NewRecorder()
	Handler()(rec, httptest.NewRequest("GET", "/metrics", nil))

	body := rec.Body.String()
	if !strings.Contains(body, "# HELP test_counter unit test") {
		t.Errorf("expected HELP line, got: %q", body)
	}
	if !strings.Contains(body, "# TYPE test_counter counter") {
		t.Errorf("expected TYPE line, got: %q", body)
	}
	if !strings.Contains(body, "test_counter 5") {
		t.Errorf("expected `test_counter 5`, got: %q", body)
	}
	if !strings.Contains(body, "go_goroutines") {
		t.Errorf("expected runtime metrics, got: %q", body)
	}
}
