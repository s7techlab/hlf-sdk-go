package errors

import (
	"fmt"
)

// todo: remove

type Error string

func (e Error) Error() string {
	return string(e)
}

type ErrNoReadyPeers struct {
	MspId string
}

func (e ErrNoReadyPeers) Error() string {
	return fmt.Sprintf("no ready peers for MspId: %s", e.MspId)
}

type ErrUnexpectedHTTPStatus struct {
	Status int
	Body   []byte
}

func (err ErrUnexpectedHTTPStatus) Error() string {
	return fmt.Sprintf("unexpected HTTP status code: %d with body %s", err.Status, string(err.Body))
}
