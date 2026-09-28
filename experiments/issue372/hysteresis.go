package main

import (
	"fmt"

	"github.com/mbrt/gmailctl/internal/engine/filter"
)

const lowThreshold = 750
const highThreshold = 900

// In the band we infer a preference, not a historical mode. A tie chooses
// normal generation, including a completely empty configuration.
func chooseHysteresis(c candidates, upstream filter.Filters, offline bool) decision {
	count := len(c.Normal)
	if count > highThreshold {
		return decision{true, fmt.Sprintf("normal count %d exceeds %d", count, highThreshold)}
	}
	if offline {
		return decision{false, fmt.Sprintf("offline export: normal count %d does not exceed %d", count, highThreshold)}
	}
	if count < lowThreshold {
		return decision{false, fmt.Sprintf("normal count %d is below %d", count, lowThreshold)}
	}
	normalCost := changeCount(upstream, c.Normal).Total()
	compactCost := changeCount(upstream, c.Compact).Total()
	return decision{
		Compact: compactCost < normalCost,
		Reason:  fmt.Sprintf("within %d–%d: normal needs %d operations; compact needs %d (ties use normal)", lowThreshold, highThreshold, normalCost, compactCost),
	}
}
