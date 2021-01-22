package fetcher

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/core/chaincode/platforms"
	"github.com/s7techlab/hlf-sdk-go/api"
)

type localFetcher struct {
	r  *platforms.Registry
	pl platforms.Platform
}

func (f *localFetcher) Fetch(ctx context.Context, id *peer.ChaincodeID) (*peer.ChaincodeDeploymentSpec, error) {
	panic(`implement me`)
	//ccBytes, err := f.r.GetDeploymentPayload(f.pl.Name(), id.Path)
	//if err != nil {
	//	return nil, errors.Wrap(err, ``)
	//}
	//
	//return &peer.ChaincodeDeploymentSpec{
	//	ChaincodeSpec: &peer.ChaincodeSpec{
	//		Type:        f.getTypeByPlatform(),
	//		ChaincodeId: id,
	//	},
	//	CodePackage: ccBytes,
	//	ExecEnv:     peer.ChaincodeDeploymentSpec_DOCKER,
	//}, nil
}

func (f *localFetcher) getTypeByPlatform() peer.ChaincodeSpec_Type {
	switch f.pl.Name() {
	case peer.ChaincodeSpec_GOLANG.String():
		return peer.ChaincodeSpec_GOLANG
	}
	return peer.ChaincodeSpec_UNDEFINED
}

func NewLocal(platform platforms.Platform) api.CCFetcher {
	return &localFetcher{r: platforms.NewRegistry(platform), pl: platform}
}
