package campaigns

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
)

// A/B testing helpers. The runner calls PickVariant(contactID) to
// deterministically bucket each recipient into one of the campaign's
// variants; sticky per contact so a resumed campaign never shifts a
// user between arms.

// PickVariant returns the template ID the given contact should receive
// for this campaign. When no Variants are configured it returns the
// fallback template. Deterministic — same (variants, contactID) always
// maps to the same arm.
func PickVariant(variantsJSON []byte, fallbackTemplateID, contactID string) string {
	var variants []string
	if len(variantsJSON) > 0 {
		_ = json.Unmarshal(variantsJSON, &variants)
	}
	if len(variants) == 0 {
		return fallbackTemplateID
	}
	// SHA-256 truncated to 8 bytes → uint64 → mod len(variants). SHA-256
	// gives us an even distribution across arms even for consecutive
	// UUID contact IDs (a plain hash/fnv would too, but this is boring
	// and correct).
	sum := sha256.Sum256([]byte(contactID))
	idx := binary.BigEndian.Uint64(sum[:8]) % uint64(len(variants))
	return variants[idx]
}

// BumpVariantStat increments one counter for one variant in the given
// stats JSON blob and returns the updated blob. The caller writes it
// back to Campaign.VariantStats in the same transaction that increments
// the campaign-level counter, so per-variant and per-campaign totals
// never drift.
func BumpVariantStat(currentJSON []byte, variantTemplateID, field string) []byte {
	stats := map[string]VariantStat{}
	if len(currentJSON) > 0 {
		_ = json.Unmarshal(currentJSON, &stats)
	}
	s := stats[variantTemplateID]
	switch field {
	case "sent":
		s.Sent++
	case "delivered":
		s.Delivered++
	case "read":
		s.Read++
	case "clicked":
		s.Clicked++
	}
	stats[variantTemplateID] = s
	b, _ := json.Marshal(stats)
	return b
}
