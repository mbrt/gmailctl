package cmd

import "github.com/pkg/errors"

// HasUserHelp returns true if an error has a help message for the user.
func HasUserHelp(err error) bool {
	uErr, ok := errors.Cause(err).(userHelp)
	return ok && uErr.Help() != ""
}

// UserError wraps the given error and makes it into a not found one
func UserError(err error, help string) error {
	if err == nil {
		return nil
	}
	return userError{err, help}
}

// GetUserHelp returns the user help associated with an error.
func GetUserHelp(err error) string {
	uErr, ok := errors.Cause(err).(userHelp)
	if ok {
		return uErr.Help()
	}
	return ""
}

type userHelp interface {
	Help() string
}

type userError struct {
	error
	help string
}

func (e userError) Error() string { return e.error.Error() }
func (e userError) Help() string  { return e.help }
