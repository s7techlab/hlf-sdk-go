package block

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"

	"github.com/s7techlab/hlf-sdk-go/proto/block"
)

func ParseTxActions(txActions []*peer.TransactionAction) ([]*block.TransactionAction, error) {
	var parsedTxActions []*block.TransactionAction

	for _, action := range txActions {
		txAction, err := ParseTxAction(action)
		if err != nil {
			return nil, fmt.Errorf("parse transaction action: %w", err)
		}
		parsedTxActions = append(parsedTxActions, txAction)
	}

	return parsedTxActions, nil
}

func ParseTxAction(txAction *peer.TransactionAction) (*block.TransactionAction, error) {
	sigHeader, err := protoutil.UnmarshalSignatureHeader(txAction.Header)
	if err != nil {
		return nil, fmt.Errorf("unmarshal signature header: %w", err)
	}

	creator, err := protoutil.UnmarshalSerializedIdentity(sigHeader.Creator)
	if err != nil {
		return nil, fmt.Errorf("unmarshal transaction creator: %w", err)
	}

	actionPayload, err := protoutil.UnmarshalChaincodeActionPayload(txAction.Payload)
	if err != nil {
		return nil, fmt.Errorf("unmarshal chaincode action from action payload: %w", err)
	}

	ccEndorserAction, err := ParseChaincodeEndorsedAction(actionPayload)
	if err != nil {
		return nil, fmt.Errorf("parse chaincode endorsed action: %w", err)
	}

	chaincodeProposalPayload, err := ParseChaincodeProposalPayload(actionPayload)
	if err != nil {
		return nil, fmt.Errorf("parse chaincode proposal payload: %w", err)
	}

	// because there is no cc version in peer.ChaincodeInvocationSpec
	if chaincodeProposalPayload.Input.ChaincodeSpec == nil {
		chaincodeProposalPayload.Input.ChaincodeSpec = &peer.ChaincodeSpec{}
	}

	if chaincodeProposalPayload.Input.ChaincodeSpec.ChaincodeId == nil {
		chaincodeProposalPayload.Input.ChaincodeSpec.ChaincodeId = &peer.ChaincodeID{}
	}
	if ccEndorserAction.ProposalResponsePayload.Extension.ChaincodeId != nil {
		chaincodeProposalPayload.Input.ChaincodeSpec.ChaincodeId.Version = ccEndorserAction.ProposalResponsePayload.Extension.ChaincodeId.Version
	}

	return &block.TransactionAction{
		Header: &block.SignatureHeader{
			Creator: creator,
			Nonce:   sigHeader.Nonce,
		},
		Payload: &block.ChaincodeActionPayload{
			ChaincodeProposalPayload: chaincodeProposalPayload,
			Action:                   ccEndorserAction,
		},
	}, nil
}

func ParseChaincodeProposalPayload(actionPayload *peer.ChaincodeActionPayload) (*block.ChaincodeProposalPayload, error) {
	chaincodeProposalPayload, err := protoutil.UnmarshalChaincodeProposalPayload(actionPayload.ChaincodeProposalPayload)
	if err != nil {
		return nil, fmt.Errorf("unmarshal chaincode proposal from action payload: %w", err)
	}

	input, err := protoutil.UnmarshalChaincodeInvocationSpec(chaincodeProposalPayload.Input)
	if err != nil {
		return nil, fmt.Errorf("unmarshal chaincode invocation spec from action payload: %w", err)
	}

	return &block.ChaincodeProposalPayload{
		Input:        input,
		TransientMap: chaincodeProposalPayload.TransientMap,
	}, nil
}

func ParseChaincodeEndorsedAction(actionPayload *peer.ChaincodeActionPayload) (*block.ChaincodeEndorsedAction, error) {
	proposalResponsePayload, err := protoutil.UnmarshalProposalResponsePayload(actionPayload.Action.ProposalResponsePayload)
	if err != nil {
		return nil, fmt.Errorf("unmarshal chaincode proposal response proposal: %w", err)
	}

	chaincodeAction, err := protoutil.UnmarshalChaincodeAction(proposalResponsePayload.Extension)
	if err != nil {
		return nil, fmt.Errorf("unmarshal chaincode action from proposal extention: %w", err)
	}

	txReadWriteSet, err := ParseTxReadWriteSet(chaincodeAction)
	if err != nil {
		return nil, fmt.Errorf("parse tx read write set from chaincode action: %w", err)
	}

	events, err := protoutil.UnmarshalChaincodeEvents(chaincodeAction.Events)
	if err != nil {
		return nil, fmt.Errorf("unmarshal cc event from chaincode action: %w", err)
	}

	var endorsements []*block.Endorsement
	for _, endorsement := range actionPayload.Action.Endorsements {
		endorser, err := protoutil.UnmarshalSerializedIdentity(endorsement.Endorser)
		if err != nil {
			return nil, fmt.Errorf("unmarshal transaction endorser: %w", err)
		}

		endorsements = append(endorsements, &block.Endorsement{
			Endorser:  endorser,
			Signature: endorsement.Signature,
		})
	}

	return &block.ChaincodeEndorsedAction{
		ProposalResponsePayload: &block.ProposalResponsePayload{
			ProposalHash: proposalResponsePayload.ProposalHash,
			Extension: &block.ChaincodeAction{
				Results:     txReadWriteSet,
				Events:      events,
				Response:    chaincodeAction.Response,
				ChaincodeId: chaincodeAction.ChaincodeId,
			},
		},
		Endorsement: endorsements,
	}, nil
}

func ParseTxReadWriteSet(chaincodeAction *peer.ChaincodeAction) (*block.TxReadWriteSet, error) {
	txReadWriteSet := &rwset.TxReadWriteSet{}
	if err := proto.Unmarshal(chaincodeAction.Results, txReadWriteSet); err != nil {
		return nil, fmt.Errorf("unmarshal txReadWriteSet from cc action result: %w", err)
	}

	var nsReadWriteSets []*block.NsReadWriteSet
	for _, nsRWset := range txReadWriteSet.NsRwset {
		kvRWSet := &kvrwset.KVRWSet{}
		if err := proto.Unmarshal(nsRWset.Rwset, kvRWSet); err != nil {
			return nil, fmt.Errorf("unmarshal kvReadWriteSet from nsRWSet: %w", err)
		}

		var collectionHashedRwset []*block.CollectionHashedReadWriteSet
		for _, hashedRwsetItem := range nsRWset.CollectionHashedRwset {
			hashedRwset := &kvrwset.HashedRWSet{}
			if err := proto.Unmarshal(hashedRwsetItem.HashedRwset, hashedRwset); err != nil {
				return nil, fmt.Errorf("unmarshal HashedRWset from collection hashedRWSet: %w", err)
			}

			collectionHashedRwset = append(collectionHashedRwset, &block.CollectionHashedReadWriteSet{
				CollectionName: hashedRwsetItem.CollectionName,
				HashedRwset:    hashedRwset,
				PvtRwsetHash:   hashedRwsetItem.PvtRwsetHash,
			})
		}

		nsReadWriteSets = append(nsReadWriteSets, &block.NsReadWriteSet{
			Namespace:             nsRWset.Namespace,
			Rwset:                 kvRWSet,
			CollectionHashedRwset: collectionHashedRwset,
		})
	}

	return &block.TxReadWriteSet{
		DataModel: txReadWriteSet.DataModel.String(),
		NsRwset:   nsReadWriteSets,
	}, nil
}
