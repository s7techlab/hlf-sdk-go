package peer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/hyperledger/fabric-protos-go/peer"
	lifecycleproto "github.com/hyperledger/fabric-protos-go/peer/lifecycle"

	"github.com/hyperledger/fabric/msp"

	"go.uber.org/zap"

	hlfsdkgo "github.com/s7techlab/hlf-sdk-go"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client/tx"
	peerproto "github.com/s7techlab/hlf-sdk-go/proto/peer"
	lifecycleccproto "github.com/s7techlab/hlf-sdk-go/proto/systemcc/lifecycle"
	"github.com/s7techlab/hlf-sdk-go/service/systemcc/lifecycle"
)

type (
	LifecycleChaincodeInfoClient struct {
		admin         msp.SigningIdentity // todo: is required here ?
		signer        msp.SigningIdentity
		querier       api.Querier
		fabricVersion hlfsdkgo.FabricVersion
		logger        *zap.Logger
	}

	LifecycleChaincodeManagerClient struct {
		admin         msp.SigningIdentity
		signer        msp.SigningIdentity
		invoker       api.Invoker
		fabricVersion hlfsdkgo.FabricVersion
		logger        *zap.Logger
	}
)

func (c *LifecycleChaincodeManagerClient) InstallChaincode(ctx context.Context, deploymentSpec *peer.ChaincodeDeploymentSpec) error {
	c.logger.Debug(`_lifecycle install chaincode as admin`,
		zap.Int(`code package length`, len(deploymentSpec.CodePackage)))

	_, err := lifecycle.New(c.invoker).InstallChaincode(
		tx.ContextWithSigner(ctx, c.admin),
		&lifecycleproto.InstallChaincodeArgs{ChaincodeInstallPackage: deploymentSpec.CodePackage})
	if err != nil {
		return fmt.Errorf("install lifecycle chaincode package: %w", err)
	}

	return nil
}

// GetInstalledChaincodes with info about channel instantiation
func (c *LifecycleChaincodeInfoClient) GetInstalledChaincodes(ctx context.Context) (*peerproto.Chaincodes, error) {
	installedChaincodes, err := lifecycle.New(c.invoker).QueryInstalledChaincodes(
		tx.ContextWithSigner(ctx, c.admin), &empty.Empty{})
	if err != nil {
		return nil, fmt.Errorf("get installed chaincodes fro lifecycle: %w", err)
	}

	ccs := &peerproto.Chaincodes{}
	for _, cc := range installedChaincodes.InstalledChaincodes {

		info := &peerproto.Chaincode{
			PackageId:        cc.PackageId, //cc.PackageId[len(cc.Label)+1:],
			LifecycleVersion: peerproto.LifecycleVersion_LIFECYCLE_V2,
		}
		for channelName := range cc.References {
			info.Channels = append(info.Channels, channelName)
		}

		// `_` (underscore) - delimiter between chaincode name and chaincode version

		undescorePos := strings.LastIndex(cc.Label, `_`)
		if undescorePos == -1 {
			return nil, fmt.Errorf(`chaincode label=%s: %w`,
				cc.Label, errors.New(`expect underscore in cс label`))
		}

		info.Name = cc.Label[0:undescorePos]
		info.Version = cc.Label[undescorePos+1:]

		ccs.Chaincodes = append(ccs.Chaincodes, info)
	}
	return ccs, nil
}

func (c *LifecycleChaincodeInfoClient) GetInstantiatedChaincodes(ctx context.Context, channel string) (*peerproto.Chaincodes, error) {
	c.logger.Debug(`query chaincode definitions from lifecycle`, zap.String(`channel`, channel))
	lifecycleSvc := lifecycle.New(c.querier)
	ccs, err := lifecycleSvc.QueryChaincodeDefinitions(
		tx.ContextWithSigner(ctx, c.admin),
		&lifecycleccproto.QueryChaincodeDefinitionsRequest{
			Channel: channel,
			Args:    &lifecycleproto.QueryChaincodeDefinitionsArgs{},
		})

	if err != nil {
		return nil, fmt.Errorf("query chaincode definitions from Lifecycle: %w", err)
	}

	c.logger.Debug(`query installed from lifecycle`)
	installedCcs, err := lifecycleSvc.QueryInstalledChaincodes(ctx, &empty.Empty{})
	if err != nil {
		return nil, fmt.Errorf("query installed chaincodes from lifecycle: %w", err)
	}

	m := make(map[string]string)
	for _, installedCc := range installedCcs.InstalledChaincodes {
		m[installedCc.Label] = installedCc.PackageId
	}

	chaincodes := &peerproto.Chaincodes{}
	for _, cc := range ccs.ChaincodeDefinitions {
		chaincodes.Chaincodes = append(chaincodes.Chaincodes, &peerproto.Chaincode{
			Name:             cc.Name,
			Version:          cc.Version,
			PackageId:        m[cc.Name+`_`+cc.Version],
			LifecycleVersion: peerproto.LifecycleVersion_LIFECYCLE_V2,
			Channels:         []string{channel},
		})
	}

	return chaincodes, nil
}

func (c *LifecycleChaincodeManagerClient) UpChaincode(
	ctx context.Context, chaincode *peerproto.Chaincode, upChaincode *peerproto.UpChaincodeRequest) (
	*peerproto.UpChaincodeResponse, error) {
	if chaincode == nil {
		return nil, errors.New(`installed chaincode data required`)
	}
	packageID := upChaincode.GetChaincodePackageId()
	packageSpec := upChaincode.GetChaincodePackageSpec()

	if packageID == nil && packageSpec == nil {
		return nil, ErrPackageIDOrSpecRequired
	}
	if packageSpec != nil && packageID == nil {
		packageID = packageSpec.Id
	}

	c.logger.Info(`upping _lifecycle chaincode`, zap.Reflect(`chaincode`, chaincode))

	// get committed version of chaincode to determine next Sequence number
	var ccSequence int64 = 1

	asAdminCtx := tx.ContextWithSigner(ctx, c.admin)
	asSignerCtx := tx.ContextWithSigner(ctx, c.signer)
	lifecycleSvc := lifecycle.New(c.invoker)

	c.logger.Info(`query chaincode definitions`, zap.String(`channel`, upChaincode.Channel))
	committedDefs, err := lifecycleSvc.QueryChaincodeDefinitions(
		asAdminCtx,
		&lifecycleccproto.QueryChaincodeDefinitionsRequest{
			Channel: upChaincode.Channel,
			Args:    &lifecycleproto.QueryChaincodeDefinitionsArgs{},
		})
	if err != nil {
		return nil, fmt.Errorf("query commited definitions: %w", err)
	}

	for _, committed := range committedDefs.ChaincodeDefinitions {
		if committed.Name == packageID.Name {
			if committed.Version == packageID.Version {
				return nil, fmt.Errorf("chaincode definition already commited")
			}
			ccSequence = committed.Sequence + 1
		}
	}

	c.logger.Info(`check commit readiness`,
		zap.String(`channel`, upChaincode.Channel),
		zap.String(`name`, packageID.Name),
		zap.String(`version`, packageID.Version))

	// Check whether chaincode definition was approved or not
	readiness, err := lifecycleSvc.CheckCommitReadiness(
		asAdminCtx,
		&lifecycleccproto.CheckCommitReadinessRequest{
			Channel: upChaincode.Channel,
			Args: &lifecycleproto.CheckCommitReadinessArgs{
				Sequence: ccSequence,
				Name:     packageID.Name,
				Version:  packageID.Version,
			},
		})

	if err != nil {
		return nil, fmt.Errorf("check commit readiness: %w", err)
	}

	approvals := readiness.Approvals
	curMSPId := c.signer.GetMSPIdentifier()
	approvedByMyOrg, ok := readiness.Approvals[curMSPId]
	if !ok {
		return nil, fmt.Errorf("current msp_id=%s not in chaincode approvers", curMSPId)
	}

	if !approvedByMyOrg {
		c.logger.Info(`approve chaincode definition for my org`,
			zap.String(`package id`, chaincode.PackageId),
			zap.Int64(`sequence`, ccSequence))

		_, err = lifecycleSvc.ApproveChaincodeDefinitionForMyOrg(
			asAdminCtx,
			&lifecycleccproto.ApproveChaincodeDefinitionForMyOrgRequest{
				Channel: upChaincode.Channel,
				Args: &lifecycleproto.ApproveChaincodeDefinitionForMyOrgArgs{
					Sequence: ccSequence,
					Name:     packageID.Name,
					Version:  packageID.Version,
					Source: &lifecycleproto.ChaincodeSource{
						Type: &lifecycleproto.ChaincodeSource_LocalPackage{
							LocalPackage: &lifecycleproto.ChaincodeSource_Local{
								PackageId: chaincode.PackageId,
							},
						},
					},
				},
			})

		if err != nil {
			return nil, fmt.Errorf("approve chaincode definition for my org: %w", err)
		}
		approvals[curMSPId] = true
	}

	result := &peerproto.UpChaincodeResponse{
		Chaincode: &peerproto.Chaincode{
			Name:             packageID.Name,
			Version:          packageID.Version,
			PackageId:        chaincode.PackageId,
			LifecycleVersion: peerproto.LifecycleVersion_LIFECYCLE_V2,
			Channels:         make([]string, 0),
		},
		Approvals: approvals,
	}

	// Trying to commit chaincode anyway
	c.logger.Info(`commit chaincode definition`,
		zap.String(`name`, packageID.Name),
		zap.String(`version`, packageID.Version),
		zap.Int64(`sequence`, ccSequence))

	_, commitErr := lifecycleSvc.CommitChaincodeDefinition(
		asSignerCtx,
		&lifecycleccproto.CommitChaincodeDefinitionRequest{
			Channel: upChaincode.Channel,
			Args: &lifecycleproto.CommitChaincodeDefinitionArgs{
				Sequence: ccSequence,
				Name:     packageID.Name,
				Version:  packageID.Version,
			},
		})

	if commitErr != nil {
		result.Committed = false
		result.CommitErr = commitErr.Error()
	} else {
		result.Committed = true
	}

	info, err := lifecycleSvc.QueryInstalledChaincode(asAdminCtx, &lifecycleproto.QueryInstalledChaincodeArgs{
		PackageId: chaincode.PackageId,
	})
	if err != nil {
		return nil, fmt.Errorf("get installed chaincode: %w", err)
	}
	for channelName := range info.References {
		result.Chaincode.Channels = append(result.Chaincode.Channels, channelName)
	}

	return result, nil
}
