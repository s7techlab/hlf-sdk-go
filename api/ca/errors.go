package ca

import (
	"fmt"
	"strings"
)

type ResponseError struct {
	Errors   []ResponseMessage
	Messages []ResponseMessage
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("CA response error messages: %s", e.joinErrors())
}

func (e *ResponseError) joinErrors() string {
	mes := make([]string, len(e.Errors))
	for i, m := range e.Errors {
		mes[i] = m.Message
	}

	return strings.Join(mes, `,`)
}
