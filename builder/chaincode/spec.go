package chaincode

import "github.com/hyperledger/fabric-protos-go/peer"

func Spec(cid *peer.ChaincodeID) *peer.ChaincodeSpec {
	return &peer.ChaincodeSpec{
		Type:        peer.ChaincodeSpec_GOLANG,
		ChaincodeId: cid,
	}
}
