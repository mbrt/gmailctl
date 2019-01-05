package v2alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyFilter(t *testing.T) {
	f := FilterNode{}
	assert.True(t, f.Empty())

	f = FilterNode{
		To: "foo",
	}
	assert.False(t, f.Empty())

	f = FilterNode{
		Not: &FilterNode{
			Cc: "me",
		},
	}
	assert.False(t, f.Empty())
}

func TestValidFilter(t *testing.T) {
	fnames := NamesSet{"f": struct{}{}}
	filters := []FilterNode{
		{To: "me"},
		{Cc: "me"},
		{Not: &FilterNode{To: "me"}},
		{And: []FilterNode{{To: "me"}, {From: "pippo"}}},
		{Or: []FilterNode{{To: "me"}, {From: "pippo"}}},
		{RefName: "f"},
	}

	for _, f := range filters {
		if err := f.Valid(fnames); err != nil {
			t.Errorf("expected filter '%+v' to be valid, got error: %v", f, err)
		}
	}
}

func TestInvalidFilter(t *testing.T) {
	fnames := NamesSet{"f": struct{}{}}
	filters := []FilterNode{
		{},
		{To: "me", Cc: "foo"},
		{Not: &FilterNode{}},
		{And: []FilterNode{}},
		{And: []FilterNode{{To: "me"}}},
		{To: "me", RefName: "foo"},
		{RefName: "g"},
	}

	for _, f := range filters {
		if err := f.Valid(fnames); err == nil {
			t.Errorf("expected filter '%+v' to be invalid, got no error", f)
		}
	}
}
