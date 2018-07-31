package api

import (
	"fmt"
	"strings"

	"github.com/cloudflare/cfssl/api"
	"github.com/pkg/errors"
)

var (
	ErrEmptyConfig = errors.New(`empty core configuration`)
)

type CAResponseError struct {
	Errors   []api.ResponseMessage
	Messages []api.ResponseMessage
}

func (e *CAResponseError) Error() string {
	return fmt.Sprintf("CA response error messages: %s", e.joinMessages())
}

func (e *CAResponseError) joinMessages() string {
	mes := make([]string, len(e.Messages))
	for i, m := range e.Messages {
		mes[i] = m.Message
	}

	return strings.Join(mes, `,`)
}
