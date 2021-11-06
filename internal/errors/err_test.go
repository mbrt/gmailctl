package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type errType struct {
	a int
}

func (errType) Error() string { return "errType" }

func TestAnnotated(t *testing.T) {
	err1 := errors.New("error1")
	err2 := errType{3}
	wrapped := WithCause(err2, err1)

	// The error is both err1 and err2
	assert.Equal(t, "error1: errType", wrapped.Error())
	assert.True(t, errors.Is(wrapped, err1))
	assert.True(t, errors.Is(wrapped, err2))

	// Contents are preserved.
	var et errType
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
