package ca

import (
	"fmt"
	"strings"
)

type ResponseError struct {
	Errors   []ResponseMessage
	Messages []ResponseMessage
}

func (err ResponseError) Error() string {
	return fmt.Sprintf("CA response error messages: %s", err.joinErrors())
}

func (err ResponseError) joinErrors() string {
	mes := make([]string, len(err.Errors))
	for i, m := range err.Errors {
		mes[i] = m.Message
	}

	return strings.Join(mes, `,`)
}
