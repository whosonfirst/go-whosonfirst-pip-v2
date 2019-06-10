// Package warning provides a simple way to handle errors that should not stop
// execution (return err), but rather continue.
//
// Common Go idiom is this:
//
// 	if err != nil {
// 		return err
// 	}
//
// But what if you wanted to distinguish between error that ends execution and
// error that should just be logged?
// This package provides you with just that.
//
//	if err != nil && warning.IsWarning(err) {
//		// This is executed only if err is not a warning.
//		return
//	}
//
// It also works well with https://github.com/pkg/errors and https://github.com/hashicorp/go-multierror.
package warning

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
)

// Warning type wraps error interface. This type is used to represent error that
// should not cause stopping of execution, but just be logged and continue.
type Warning struct {
	error
}

// New creates new Warning from a message.
func New(message string) error {
	return Warning{fmt.Errorf("%s", message)}
}

// IsWarning returns true if given error is Warning type.
func IsWarning(err error) bool {
	switch v := err.(type) {
	case Warning:
		return true
	case *multierror.Error:
		// In case of multierror, we have to iterate over all of wrapped errors
		// and return true only if ALL of errors are Warnings.
		for _, merr := range v.WrappedErrors() {
			if !IsWarning(merr) {
				return false
			}
		}
		return true
	}
	return false
}

// Wrap wraps any error into Warning.
func Wrap(err error) error {
	// Do not wrap nil.
	if err == nil {
		return nil
	}
	return Warning{err}
}

// Cause returns the underlying cause of Warning.
func (w *Warning) Cause() error {
	return w.error
}
