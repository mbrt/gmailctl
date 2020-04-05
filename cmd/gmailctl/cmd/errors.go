package cmd

import "errors"

// HasUserHelp returns true if an error has a help message for the user.
func HasUserHelp(err error) bool {
	var uErr userHelp
	if errors.As(err, &uErr) {
		return uErr.Help() != ""
	}
	return false
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
	var uErr userHelp
	if errors.As(err, &uErr) {
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
