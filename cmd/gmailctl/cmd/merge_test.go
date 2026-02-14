package cmd

import (
	"testing"
)

func TestMergeStrategyValues(t *testing.T) {
	if StrategyLocal != "local" {
		t.Errorf("StrategyLocal = %q, want 'local'", StrategyLocal)
	}
	if StrategyGmail != "gmail" {
		t.Errorf("StrategyGmail = %q, want 'gmail'", StrategyGmail)
	}
	if StrategyPrompt != "" {
		t.Errorf("StrategyPrompt = %q, want ''", StrategyPrompt)
	}
}

func TestValidStrategies(t *testing.T) {
	strategies := ValidStrategies()
	if len(strategies) != 2 {
		t.Fatalf("ValidStrategies() returned %d strategies, want 2", len(strategies))
	}
	// Check both are present
	hasLocal, hasGmail := false, false
	for _, s := range strategies {
		if s == "local" {
			hasLocal = true
		}
		if s == "gmail" {
			hasGmail = true
		}
	}
	if !hasLocal {
		t.Error("ValidStrategies() missing 'local'")
	}
	if !hasGmail {
		t.Error("ValidStrategies() missing 'gmail'")
	}
}
