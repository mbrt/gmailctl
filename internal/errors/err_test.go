package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type someError struct {
	a int
}

func (someError) Error() string { return "someError" }

func TestAnnotated(t *testing.T) {
	err1 := errors.New("error1")
	err2 := someError{3}
	wrapped := WithCause(err2, err1)

	// The error is both err1 and err2
	assert.Equal(t, "error1: someError", wrapped.Error())
	assert.True(t, errors.Is(wrapped, err1))
	assert.True(t, errors.Is(wrapped, err2))

	// Contents are preserved.
	var et someError
	assert.True(t, errors.As(wrapped, &et))
	assert.Equal(t, err2, et)

	// Same properties with the errors inverted.
	wrapped = WithCause(err1, err2)
	assert.True(t, errors.Is(wrapped, err1))
	assert.True(t, errors.Is(wrapped, err2))
	assert.True(t, errors.As(wrapped, &et))
	assert.Equal(t, err2, et)
}

func TestErrWithDetails(t *testing.T) {
	err1 := errors.New("err1")
	err2 := WithDetails(
		fmt.Errorf("err2: %w", err1),
		"second descr\nmultiline",
		"another\ndescr",
	)
	err3 := WithDetails(
		fmt.Errorf("err3: %w", err2),
		"third descr\nmultiline\nmultiline again")
	err4 := fmt.Errorf("err4: %w", err3)
	assert.Equal(t, "err4: err3: err2: err1", fmt.Sprintf("%v", err4))
	details := `
  - third descr
    multiline
    multiline again
  - second descr
    multiline
  - another
    descr`
	assert.Equal(t, details, Details(err4))
}

func TestCombine(t *testing.T) {
	err1 := errors.New("err1")
	err2 := errors.New("err2")

	// Combine should ignore nils.
	assert.NoError(t, Combine())
	assert.NoError(t, Combine(nil))
	assert.NoError(t, Combine(nil, nil))
	assert.EqualError(t, Combine(err1, nil), err1.Error())

	// Combine two.
	assert.ElementsMatch(t, Errors(Combine(err1, err2)), []error{err1, err2})
	assert.ElementsMatch(t, Errors(Combine(err1, nil, err2)), []error{err1, err2})

	// Nesting.
	err3 := Combine(err1, err2)
	err4 := Combine(err1, err2, err3, nil)
	assert.ElementsMatch(t, Errors(err4), []error{err1, err1, err2, err2})

	// Print.
	assert.EqualError(t, err4, "multiple errors (4); sample: err1")
	verbose := `multiple errors (4):
- err1
- err2
- err1
- err2`
	assert.Equal(t, fmt.Sprintf("%+v", err4), verbose)
}
