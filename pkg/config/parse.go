package config

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	cfgv1 "github.com/mbrt/gmailctl/pkg/config/v1alpha1"
)

// ParseFile takes a path and returns the parsed config file.
func ParseFile(path string) (cfgv1.Config, error) {
	/* #nosec */
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return cfgv1.Config{}, NotFoundError(err)
	}

	var res cfgv1.Config
	err = yaml.UnmarshalStrict(b, &res)
	return res, err
}

// IsNotFound returns true if an error is related to a file not found
func IsNotFound(err error) bool {
	nfErr, ok := errors.Cause(err).(notFound)
	return ok && nfErr.NotFound()
}

// NotFoundError wraps the given error and makes it into a not found one
func NotFoundError(err error) error {
	if err == nil {
		return nil
	}
	return notFoundError{err}
}

type notFound interface {
	NotFound() bool
}

type notFoundError struct {
	error
}

func (e notFoundError) Error() string  { return e.error.Error() }
func (e notFoundError) NotFound() bool { return true }
