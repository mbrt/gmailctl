package label

// MergeLabels merges two label sets with Gmail-wins strategy for conflicts.
// Returns a merged label set where:
// - Labels only in Gmail are added
// - Labels only in local are kept
// - Labels in both use Gmail's version (preserves Gmail color settings)
func MergeLabels(local, gmail Labels) Labels {
	// Build a map of local labels by name
	localByName := make(map[string]Label)
	for _, l := range local {
		localByName[l.Name] = l
	}

	// Build a map of Gmail labels by name
	gmailByName := make(map[string]Label)
	for _, l := range gmail {
		gmailByName[l.Name] = l
	}

	// Result starts with all Gmail labels (Gmail wins on overlaps)
	result := make(Labels, 0, len(gmail)+len(local))
	result = append(result, gmail...)

	// Add any local-only labels (name not in Gmail)
	for name, localLabel := range localByName {
		if _, existsInGmail := gmailByName[name]; !existsInGmail {
			result = append(result, localLabel)
		}
	}

	return result
}
