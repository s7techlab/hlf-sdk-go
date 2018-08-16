package api

import (
	"fmt"

	"github.com/pkg/errors"
)

var (
	ErrEmptyConfig = errors.New(`empty core configuration`)
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
