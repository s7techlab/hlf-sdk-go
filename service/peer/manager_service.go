package peer

import (
	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/service/ccpackage"
)

type ManagerService struct {
	packages ccpackage.PackageServiceServer
	timeouts *ManageTimeouts
	logger   *zap.Logger
}
