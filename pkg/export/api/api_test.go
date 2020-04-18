package api

import (
	"reflect"
	"testing"

	gmailv1 "google.golang.org/api/gmail/v1"
)

func TestUnknownFilterFields(t *testing.T) {
	// Test that new versions of the API don't add fields that we don't know
	// about.
	// This makes sure we acknowledge the existence of new fields so that we
	// don't screw up user filters on updates.
	tests := []struct {
		value       interface{}
		knownFields map[string]bool
	}{
		{
			value:       gmailv1.FilterAction{},
			knownFields: knownActionFields,
		},
		{
			value:       gmailv1.FilterCriteria{},
			knownFields: knownCriteriaFields,
		},
	}

	for _, tc := range tests {
		tp := reflect.TypeOf(tc.value)
		t.Run(tp.Name(), func(t *testing.T) {
			for i := 0; i < tp.NumField(); i++ {
				field := tp.Field(i)
				if !tc.knownFields[field.Name] {
					t.Errorf("Field %q is unknown", field.Name)
				}
			}
		})
	}
}
