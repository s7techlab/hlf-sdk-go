package peer

import (
	"time"
)

// ManageTimeouts for peer operation
type (
	ManageTimeouts struct {
		ChaincodeCreatePackage time.Duration // time to build package
		Endorsement            time.Duration // max time of any endorsement
		ChaincodeInstall       time.Duration
		ChaincodeUp            time.Duration
		ChannelJoin            time.Duration
	}

	ReadTimeouts struct {
		// ReadRequest timeout for requesting config (channel config, channel list, chaincode  etc)
		ReadRequest       time.Duration
		CheckConnectivity time.Duration
	}
)

var (
	DefaultManageTimeouts = &ManageTimeouts{
		ChaincodeCreatePackage: 300 * time.Second,
		Endorsement:            300 * time.Second,

		ChaincodeInstall: 300 * time.Second,
		ChaincodeUp:      120 * time.Second,

		ChannelJoin: 10 * time.Second,
	}

	DefaultReadTimeouts = &ReadTimeouts{
		ReadRequest:       10 * time.Second,
		CheckConnectivity: 5 * time.Second,
	}
)
