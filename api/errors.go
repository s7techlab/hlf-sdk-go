package api

import (
	"fmt"
)

const (
	ErrEmptyConfig         = Error(`empty core configuration`)
	ErrInvalidPEMStructure = Error(`invalid PEM structure`)
)

type MultiError struct {
	Errors []error
}

func (e *MultiError) Error() string {
	errStr := fmt.Sprintf("next errors occurred:\n")
	for _, err := range e.Errors {
		errStr += fmt.Sprintf("%s\n", err.Error())
	}
	return errStr
}

func (e *MultiError) Add(err error) {
	e.Errors = append(e.Errors, err)
}

type ErrUnexpectedHTTPStatus struct {
	Status int
	Body   []byte
}

func (err ErrUnexpectedHTTPStatus) Error() string {
	return fmt.Sprintf("unexpected HTTP status code: %d with body %s", err.Status, string(err.Body))
}

type Error string

func (e Error) Error() string {
	return string(e)
}
