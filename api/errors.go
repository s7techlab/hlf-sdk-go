package api

import (
	"github.com/pkg/errors"
)

var (
	ErrEmptyConfig = errors.New(`empty core configuration`)
)
