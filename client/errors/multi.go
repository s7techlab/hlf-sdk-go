package errors

import (
	"fmt"
)

type MultiError struct {
	Errors []error
}

func (e *MultiError) Error() string {
	errStr := "next errors occurred:\n"
	for _, err := range e.Errors {
		errStr += fmt.Sprintf("%s\n", err.Error())
	}
	return errStr
}

func (e *MultiError) Add(err error) {
	e.Errors = append(e.Errors, err)
}
