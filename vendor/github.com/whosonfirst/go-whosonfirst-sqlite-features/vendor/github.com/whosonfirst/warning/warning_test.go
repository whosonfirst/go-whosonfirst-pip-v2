package warning_test

import (
	"fmt"
	"testing"

	"warning"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// TestNew tests if New function returns Warning.
func TestNew(t *testing.T) {
	err := warning.New("this is wrong")
	if !warning.IsWarning(err) {
		t.Errorf("new does not return Warning")
	}
}

// TestUseCase tests basic use case.
func TestUseCase(t *testing.T) {
	err := warning.Wrap(fmt.Errorf("computation error"))
	// If error happens, stop all other code execution. But if this error is a
	// Warning, we want to continue - skip return.
	if err != nil && !warning.IsWarning(err) {
		t.Errorf("this should not execute")
		return
	}
}

// TestNil tests if nil is not a warning.
func TestNil(t *testing.T) {
	var err error
	if warning.IsWarning(err) {
		t.Errorf("nil is not a warning")
	}
}

// TestErrorIsNotWarning tests if common error is not a Warning.
func TestErrorIsNotWarning(t *testing.T) {
	err := fmt.Errorf("common error")
	if err != nil && !warning.IsWarning(err) {
		return
	}
	t.Errorf("this should not execute")
}

// TestCauseWrap tests if basic errors.Wrap usage can be used with Warning.
func TestCauseWrap(t *testing.T) {
	err := warning.Wrap(errors.Wrap(fmt.Errorf("validation error"), "unable to parse"))
	if !warning.IsWarning(err) {
		t.Errorf("this should not execute")
	}
}

// TestMultierrSingleWarning tests if multierror with single warning behaves like a warning.
func TestMultierrSingleWarning(t *testing.T) {
	var multierr *multierror.Error
	multierr = multierror.Append(multierr, warning.New("warning"))

	if !warning.IsWarning(multierr.ErrorOrNil()) {
		t.Errorf("this should not execute")
	}
}

// TestMultierrMultipleWarnings tests if multierror with multiple warnings behaves like a warning.
func TestMultierrMultipleWarnings(t *testing.T) {
	var multierr *multierror.Error
	multierr = multierror.Append(multierr, warning.New("warning"))
	multierr = multierror.Append(multierr, warning.New("another warning"))

	if !warning.IsWarning(multierr.ErrorOrNil()) {
		t.Errorf("this should not execute")
	}
}

// TestMultierrNonWarning tests if multierror with warning and error don't behave like a warning.
func TestMultierrNonWarning(t *testing.T) {
	var multierr *multierror.Error
	multierr = multierror.Append(multierr, warning.New("warning"))
	multierr = multierror.Append(multierr, errors.New("another warning"))

	if warning.IsWarning(multierr.ErrorOrNil()) {
		t.Errorf("this should not execute")
	}
}
