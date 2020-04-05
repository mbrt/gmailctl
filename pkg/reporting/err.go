package reporting

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Prettify returns a readable string representing the object.
func Prettify(o interface{}, compact bool) string {
	var (
		b   []byte
		err error
	)
	if compact {
		b, err = json.Marshal(o)
	} else {
		b, err = json.MarshalIndent(o, "", "  ")
	}
	if err != nil {
		return fmt.Sprintf("(invalid) %+v", o)
	}
	return string(b)
}

// AnnotateErr annotates a symptom error with a cause.
//
// Both errors can be discovered by the Is and As methods.
func AnnotateErr(cause, symptom error) error {
	return errAnnotated{
		cause:   cause,
		symptom: symptom,
	}
}

type errAnnotated struct {
	cause   error
	symptom error
}

func (e errAnnotated) Error() string {
	return fmt.Sprintf("%s: %s", e.cause, e.symptom)
}

func (e errAnnotated) Unwrap() error {
	return e.cause
}

func (e errAnnotated) Is(target error) bool {
	return errors.Is(e.symptom, target) || errors.Is(e.cause, target)
}

func (e errAnnotated) As(target interface{}) bool {
	if errors.As(e.symptom, target) {
		return true
	}
	return errors.As(e.cause, target)
}
