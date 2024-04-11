package peer

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/policydsl"
	"github.com/hyperledger/fabric/msp"
	"go.uber.org/zap"

	hlfsdkgo "github.com/s7techlab/hlf-sdk-go"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
	"github.com/s7techlab/hlf-sdk-go/service/systemcc/cscc"
	"github.com/s7techlab/hlf-sdk-go/service/systemcc/lscc"
)

type (
	LSCCChaincodeInfoClient struct {
		admin         msp.SigningIdentity
		signer        msp.SigningIdentity
		invoker       api.Invoker
		fabricVersion hlfsdkgo.FabricVersion
		logger        *zap.Logger
	}

	LSCCChaincodeManagerClient struct {
		admin         msp.SigningIdentity
		signer        msp.SigningIdentity
		invoker       api.Invoker
		fabricVersion hlfsdkgo.FabricVersion
		logger        *zap.Logger
	}
)

// InstallChaincode install chaincode over LSCC
func (c *LSCCChaincodeManagerClient) InstallChaincode(ctx context.Context, deploymentSpec *peer.ChaincodeDeploymentSpec) error {
	c.logger.Debug(`LSCC install chaincode as admin`,
		zap.Int(`code package length`, len(deploymentSpec.CodePackage)))

	_, err := lscc.New(c.invoker).Install(tx.ContextWithSigner(ctx, c.admin), deploymentSpec)
	return err
}

func (c *LSCCChaincodeInfoClient) GetInstantiatedChaincodes(ctx context.Context, channel string) (*Chaincodes, error) {
	c.logger.Debug(`get instantiated chaincodes from lscc`, zap.String(`channel`, channel))
	instChaincodes, err := lscc.New(c.invoker).GetChaincodes(
		tx.ContextWithSigner(ctx, c.admin),
		&lscc.GetChaincodesRequest{Channel: channel})
	if err != nil {
		return nil, fmt.Errorf("get chaincodes in channel=%s from LSCC: %w", channel, err)
	}

	chaincodes := &Chaincodes{}
	for _, cc := range instChaincodes.Chaincodes {
		chaincodes.Chaincodes = append(chaincodes.Chaincodes, &Chaincode{
			Name:             cc.Name,
			Version:          cc.Version,
			PackageId:        hex.EncodeToString(cc.Id),
			LifecycleVersion: LifecycleVersion_LIFECYCLE_V1,
			Channels:         []string{channel},
		})
	}

	return chaincodes, nil
}

// GetInstalledChaincodes with info about channel instantiation
func (c *LSCCChaincodeInfoClient) GetInstalledChaincodes(ctx context.Context) (*Chaincodes, error) {
	ctxAsAdmin := tx.ContextWithSigner(ctx, c.admin)
	lsccSvc := lscc.New(c.invoker)

	installedChaincodes, err := lsccSvc.GetInstalledChaincodes(ctxAsAdmin, &empty.Empty{})
	if err != nil {
		return nil, fmt.Errorf("get installed chaincodes from LSCC: %w", err)
	}

	channelsList, err := cscc.New(c.invoker, c.fabricVersion).GetChannels(ctxAsAdmin, &empty.Empty{})
	if err != nil {
		return nil, fmt.Errorf("get channels: %w", err)
	}

	ccInChannel := make(map[string][]string)
	// get info about chaincode instantiation in channel
	for _, ch := range channelsList.Channels {
		channelChaincodes, err := lsccSvc.GetChaincodes(ctxAsAdmin, &lscc.GetChaincodesRequest{Channel: ch.ChannelId})
		if err != nil {
			return nil, fmt.Errorf("get chaincodes in channel=%s from LSCC: %w", ch.ChannelId, err)
		}

		for _, cc := range channelChaincodes.Chaincodes {
			ccID := fmt.Sprintf("%v:%v", cc.Name, cc.Version)
			ccInChannel[ccID] = append(ccInChannel[ccID], ch.ChannelId)
		}
	}

	ccs := &Chaincodes{}
	for _, cc := range installedChaincodes.Chaincodes {
		ccs.Chaincodes = append(ccs.Chaincodes, &Chaincode{
			Name:             cc.Name,
			Version:          cc.Name,
			PackageId:        "",
			LifecycleVersion: LifecycleVersion_LIFECYCLE_V1,
			Channels:         ccInChannel[fmt.Sprintf("%v:%v", cc.Name, cc.Version)],
		})
	}

	return ccs, nil
}

func (c *LSCCChaincodeManagerClient) UpChaincode(ctx context.Context, chaincode *Chaincode, depSpec *peer.ChaincodeDeploymentSpec, upChaincode *UpChaincodeRequest) (*UpChaincodeResponse, error) {
	c.logger.Info(`up LSCC chaincode`, zap.Reflect(`chaincode`, chaincode))

	depSpec = proto.Clone(depSpec).(*peer.ChaincodeDeploymentSpec)

	policy, err := policydsl.FromString(upChaincode.Policy)
	if err != nil {
		return nil, fmt.Errorf(`parse endorsement policy= %s: %w`, upChaincode.Policy, err)
	}

	if upChaincode.Input != nil {
		depSpec.ChaincodeSpec.Input = &peer.ChaincodeInput{Args: upChaincode.Input.Args}
	}

	c.logger.Info(`deploy chaincode via LSCC`,
		zap.String(`channel`, upChaincode.Channel),
		zap.Reflect(`chaincode`, depSpec.ChaincodeSpec))

	_, err = lscc.New(c.invoker).Deploy(tx.ContextWithSigner(ctx, c.admin), &lscc.DeployRequest{
		Channel:        upChaincode.Channel,
		DeploymentSpec: depSpec,
		Policy:         policy,
		Transient:      upChaincode.TransientMap,
	})
	if err != nil {
		return nil, fmt.Errorf(`deploy chaincode via LSCC: %w`, err)
	}

	chaincode.Channels = append(chaincode.Channels, upChaincode.Channel)

	return &UpChaincodeResponse{
		Chaincode: nil,
		Approvals: nil,
		Committed: true,
		CommitErr: "",
	}, nil
}
