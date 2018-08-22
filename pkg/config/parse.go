package config

import (
	"io/ioutil"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// ParseFile takes a path and returns the parsed config file.
func ParseFile(path string) (Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, NotFoundError(err)
	}

	var res Config
	err = yaml.Unmarshal(b, &res)
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
