package filter

// MergeCategories contains filters categorized for merge operations.
type MergeCategories struct {
	// Matched filters are identical in both local and Gmail (by content, ignoring ID)
	Matched Filters
	// GmailOnly filters exist in Gmail but not in local
	GmailOnly Filters
	// LocalOnly filters exist in local but not in Gmail
	LocalOnly Filters
	// Conflicts are filters with same criteria but different actions
	Conflicts []FilterConflict
}

// FilterConflict represents a filter that exists in both local and Gmail
// with the same criteria but different actions.
type FilterConflict struct {
	Local Filter
	Gmail Filter
}

// CategorizeMerge categorizes filters for interactive merge.
//
// Conflict detection algorithm:
// 1. Hash each filter by full content (criteria + actions) to find exact matches
// 2. For remaining unmatched filters, hash by criteria only
// 3. If same criteria hash exists on both sides, it's a conflict (same filter, different actions)
// 4. Otherwise, filter is gmail-only or local-only
func CategorizeMerge(local, gmail Filters) MergeCategories {
	var result MergeCategories

	// Step 1: Build content hash maps (criteria + actions)
	// Using existing hashFilter() from diff.go which hashes by content excluding ID
	localByContent := make(map[string]Filter)
	gmailByContent := make(map[string]Filter)

	for _, f := range local {
		h := hashFilter(f)
		localByContent[h.hash] = f
	}
	for _, f := range gmail {
		h := hashFilter(f)
		gmailByContent[h.hash] = f
	}

	// Step 2: Find exact matches (same content hash = identical filter)
	matchedLocalHashes := make(map[string]bool)
	matchedGmailHashes := make(map[string]bool)

	for hash, localFilter := range localByContent {
		if _, exists := gmailByContent[hash]; exists {
			result.Matched = append(result.Matched, localFilter)
			matchedLocalHashes[hash] = true
			matchedGmailHashes[hash] = true
		}
	}

	// Step 3: Build criteria-only hash maps for unmatched filters
	// This detects "same filter, different actions" conflicts
	localByCriteria := make(map[string]Filter)
	gmailByCriteria := make(map[string]Filter)

	for hash, f := range localByContent {
		if !matchedLocalHashes[hash] {
			criteriaHash := hashStruct(f.Criteria)
			localByCriteria[criteriaHash] = f
		}
	}
	for hash, f := range gmailByContent {
		if !matchedGmailHashes[hash] {
			criteriaHash := hashStruct(f.Criteria)
			gmailByCriteria[criteriaHash] = f
		}
	}

	// Step 4: Categorize remaining filters
	processedCriteria := make(map[string]bool)

	for criteriaHash, localFilter := range localByCriteria {
		if gmailFilter, exists := gmailByCriteria[criteriaHash]; exists {
			// Same criteria, different actions = conflict
			result.Conflicts = append(result.Conflicts, FilterConflict{
				Local: localFilter,
				Gmail: gmailFilter,
			})
			processedCriteria[criteriaHash] = true
		} else {
			// Only in local
			result.LocalOnly = append(result.LocalOnly, localFilter)
		}
	}

	for criteriaHash, gmailFilter := range gmailByCriteria {
		if !processedCriteria[criteriaHash] {
			// Only in Gmail
			result.GmailOnly = append(result.GmailOnly, gmailFilter)
		}
	}

	return result
}
