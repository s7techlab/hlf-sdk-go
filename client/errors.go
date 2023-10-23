package client

import (
	"github.com/s7techlab/hlf-sdk-go/client/errors"
)

const (
	ErrEmptyConfig         = errors.Error(`empty config`)
	ErrInvalidPEMStructure = errors.Error(`invalid PEM structure`)

	ErrNoPeersForMSP = errors.Error(`no peers for MSP`)
	ErrMSPNotFound   = errors.Error(`MSP not found`)
	ErrPeerNotReady  = errors.Error(`peer not ready`)
)
