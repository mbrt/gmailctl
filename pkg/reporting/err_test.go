package reporting

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type errType struct {
	a int
}

func (_ errType) Error() string { return "errType" }

func TestAnnotateErr(t *testing.T) {
	err1 := errors.New("error1")
	err2 := errType{3}
	wrapped := AnnotateErr(err1, err2)

	// The error is both err1 and err2
	assert.Equal(t, "error1: errType", wrapped.Error())
	assert.True(t, errors.Is(wrapped, err1))
	assert.True(t, errors.Is(wrapped, err2))

	// Contents are preserved.
	var et errType
	assert.True(t, errors.As(wrapped, &et))
	assert.Equal(t, err2, et)

	// Same properties with the errors inverted.
	wrapped = AnnotateErr(err2, err1)
	assert.True(t, errors.Is(wrapped, err1))
	assert.True(t, errors.Is(wrapped, err2))
	assert.True(t, errors.As(wrapped, &et))
	assert.Equal(t, err2, et)
}
