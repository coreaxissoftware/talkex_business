package paylinks

import (
	"strings"
	"testing"
)

func TestSimulatedURLShape(t *testing.T) {
	pl := &PayLink{ContactID: "abcdef01-2345-6789-abcd-ef0123456789", AmountPaise: 50000}
	got := simulatedURL(pl)
	if !strings.HasPrefix(got, "https://rzp.io/sim/") {
		t.Fatalf("bad prefix: %s", got)
	}
	if !strings.Contains(got, "abcdef01") {
		t.Fatalf("missing contact prefix: %s", got)
	}
	if !strings.HasSuffix(got, "-50000") {
		t.Fatalf("missing amount suffix: %s", got)
	}
}
