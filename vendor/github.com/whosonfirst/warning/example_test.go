package warning_test

import (
	"errors"
	"fmt"

	"warning"

	"github.com/hashicorp/go-multierror"
)

func ExampleWrap() {
	err := warning.Wrap(errors.New("something happened but I don't want to stop"))
	if err != nil && warning.IsWarning(err) {
		fmt.Println("this is a warning")
	}

	// Output: this is a warning
}

func Example() {
	myfunc := func() error {
		// Suppose more complicated function here.
		var multierr *multierror.Error
		for i := 0; i <= 10; i++ {
			// Process data but do not stop on errors, create just warnings.
			msg := fmt.Sprintf("Item %d did not complete.", i)
			multierr = multierror.Append(multierr, warning.New(msg))
		}

		return multierr.ErrorOrNil()
	}

	err := myfunc()
	if err != nil && !warning.IsWarning(err) {
		// Stop execution if error is not a warning.
		return
	}

	fmt.Println(err)

	// Output: 11 errors occurred:
	//
	// * Item 0 did not complete.
	// * Item 1 did not complete.
	// * Item 2 did not complete.
	// * Item 3 did not complete.
	// * Item 4 did not complete.
	// * Item 5 did not complete.
	// * Item 6 did not complete.
	// * Item 7 did not complete.
	// * Item 8 did not complete.
	// * Item 9 did not complete.
	// * Item 10 did not complete.
}
