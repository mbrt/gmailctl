package filter

import (
	"testing"
)

func TestCategorizeMerge_EmptyBoth(t *testing.T) {
	result := CategorizeMerge(nil, nil)
	if len(result.Matched) != 0 || len(result.GmailOnly) != 0 ||
		len(result.LocalOnly) != 0 || len(result.Conflicts) != 0 {
		t.Error("Expected all empty categories for nil inputs")
	}
}

func TestCategorizeMerge_GmailOnlyFilters(t *testing.T) {
	gmail := Filters{
		{Criteria: Criteria{From: "test@example.com"}, Action: Actions{Archive: true}},
	}
	result := CategorizeMerge(nil, gmail)
	if len(result.GmailOnly) != 1 {
		t.Errorf("Expected 1 gmail-only filter, got %d", len(result.GmailOnly))
	}
	if len(result.LocalOnly) != 0 || len(result.Matched) != 0 || len(result.Conflicts) != 0 {
		t.Error("Expected only GmailOnly to be populated")
	}
}

func TestCategorizeMerge_LocalOnlyFilters(t *testing.T) {
	local := Filters{
		{Criteria: Criteria{From: "local@example.com"}, Action: Actions{MarkRead: true}},
	}
	result := CategorizeMerge(local, nil)
	if len(result.LocalOnly) != 1 {
		t.Errorf("Expected 1 local-only filter, got %d", len(result.LocalOnly))
	}
	if len(result.GmailOnly) != 0 || len(result.Matched) != 0 || len(result.Conflicts) != 0 {
		t.Error("Expected only LocalOnly to be populated")
	}
}

func TestCategorizeMerge_MatchedFilters(t *testing.T) {
	// Same criteria AND same actions = matched (identical filter)
	filter := Filter{
		Criteria: Criteria{From: "same@example.com"},
		Action:   Actions{Archive: true},
	}
	local := Filters{filter}
	gmail := Filters{filter}

	result := CategorizeMerge(local, gmail)
	if len(result.Matched) != 1 {
		t.Errorf("Expected 1 matched filter, got %d", len(result.Matched))
	}
	if len(result.Conflicts) != 0 {
		t.Error("Identical filters should not be conflicts")
	}
}

func TestCategorizeMerge_ConflictFilters(t *testing.T) {
	// Same criteria but DIFFERENT actions = conflict
	localFilter := Filter{
		Criteria: Criteria{From: "conflict@example.com"},
		Action:   Actions{Archive: true},
	}
	gmailFilter := Filter{
		Criteria: Criteria{From: "conflict@example.com"},
		Action:   Actions{Delete: true}, // Different action
	}
	local := Filters{localFilter}
	gmail := Filters{gmailFilter}

	result := CategorizeMerge(local, gmail)
	if len(result.Conflicts) != 1 {
		t.Errorf("Expected 1 conflict, got %d", len(result.Conflicts))
	}
	if result.Conflicts[0].Local.Action.Archive != true {
		t.Error("Conflict should preserve local filter")
	}
	if result.Conflicts[0].Gmail.Action.Delete != true {
		t.Error("Conflict should preserve Gmail filter")
	}
}

func TestCategorizeMerge_MixedScenario(t *testing.T) {
	// Create a mix of all categories
	matchedFilter := Filter{
		Criteria: Criteria{From: "matched@example.com"},
		Action:   Actions{Archive: true},
	}
	localOnlyFilter := Filter{
		Criteria: Criteria{From: "localonly@example.com"},
		Action:   Actions{MarkRead: true},
	}
	gmailOnlyFilter := Filter{
		Criteria: Criteria{From: "gmailonly@example.com"},
		Action:   Actions{Star: true},
	}
	conflictLocal := Filter{
		Criteria: Criteria{From: "conflict@example.com"},
		Action:   Actions{Archive: true},
	}
	conflictGmail := Filter{
		Criteria: Criteria{From: "conflict@example.com"},
		Action:   Actions{Delete: true},
	}

	local := Filters{matchedFilter, localOnlyFilter, conflictLocal}
	gmail := Filters{matchedFilter, gmailOnlyFilter, conflictGmail}

	result := CategorizeMerge(local, gmail)

	if len(result.Matched) != 1 {
		t.Errorf("Expected 1 matched, got %d", len(result.Matched))
	}
	if len(result.LocalOnly) != 1 {
		t.Errorf("Expected 1 local-only, got %d", len(result.LocalOnly))
	}
	if len(result.GmailOnly) != 1 {
		t.Errorf("Expected 1 gmail-only, got %d", len(result.GmailOnly))
	}
	if len(result.Conflicts) != 1 {
		t.Errorf("Expected 1 conflict, got %d", len(result.Conflicts))
	}
}
