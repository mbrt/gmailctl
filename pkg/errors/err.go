package errors

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// Aliases to the standard errors package.
var (
	New = errors.New
	Is  = errors.Is
	As  = errors.As
)

// New is an alias for errors.New.

// WithCause annotates a symptom error with a cause.
//
// Both errors can be discovered by the Is and As methods.
func WithCause(symptom, cause error) error {
	return annotated{
		cause:   cause,
		symptom: symptom,
	}
}

func WithDetails(err error, details ...string) error {
	if err == nil {
		return nil
	}
	return detailed{err, details}
}

func Details(err error) string {
	var (
		buffer bytes.Buffer
		dErr   detailed
	)
	for errors.As(err, &dErr) {
		// Append all details of this error.
		for _, d := range dErr.details {
			buffer.WriteString("\n  - ")
			buffer.WriteString(strings.ReplaceAll(d, "\n", "\n    "))
		}
		// Continue down the chain.
		err = dErr.error
	}
	return buffer.String()
}

type detailed struct {
	error
	details []string
}

type annotated struct {
	cause   error
	symptom error
}

func (e annotated) Error() string {
	return fmt.Sprintf("%s: %s", e.cause, e.symptom)
}

func (e annotated) Unwrap() error {
	return e.cause
}

func (e annotated) Is(target error) bool {
	return errors.Is(e.symptom, target) || errors.Is(e.cause, target)
}

func (e annotated) As(target interface{}) bool {
	if errors.As(e.symptom, target) {
		return true
	}
	return errors.As(e.cause, target)
}
