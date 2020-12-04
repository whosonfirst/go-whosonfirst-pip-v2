package crawl

import (
	"fmt"
)

type DoneError struct{}

func (e *DoneError) Error() string {
	return "context.Context signaled Done()"
}

type CrawlError struct {
	Path    string
	Details error
}

func (e *CrawlError) Error() string {
	return e.String()
}

func (e *CrawlError) String() string {
	return fmt.Sprintf("Failed crawl for %s: %v", e.Path, e.Details)
}

type WalkError struct {
	Path    string
	Details error
}

func (e *WalkError) Error() string {
	return e.String()
}

func (e *WalkError) String() string {
	return fmt.Sprintf("Failed walk for %s: %v", e.Path, e.Details)
}

type CallbackError struct {
	Path    string
	Details error
}

func (e *CallbackError) Error() string {
	return e.String()
}

func (e *CallbackError) String() string {
	return fmt.Sprintf("Failed crawl callback for %s: %v", e.Path, e.Details)
}

func NewDoneError() *DoneError {
	return &DoneError{}
}

func NewCrawlError(path string, details error) *CrawlError {

	err := CrawlError{
		Path:    path,
		Details: details,
	}

	return &err
}

func NewWalkError(path string, details error) *WalkError {

	err := WalkError{
		Path:    path,
		Details: details,
	}

	return &err
}

func NewCallbackError(path string, details error) *CallbackError {

	err := CallbackError{
		Path:    path,
		Details: details,
	}

	return &err
}

func IsDoneError(err error) bool {

	switch err.(type) {
	case *DoneError:
		return true
	default:
		return false
	}
}

func IsCrawlError(err error) bool {

	switch err.(type) {
	case *CrawlError:
		return true
	default:
		return false
	}
}

func IsWalkError(err error) bool {

	switch err.(type) {
	case *WalkError:
		return true
	default:
		return false
	}
}

func IsCallbackError(err error) bool {

	switch err.(type) {
	case *CallbackError:
		return true
	default:
		return false
	}
}
