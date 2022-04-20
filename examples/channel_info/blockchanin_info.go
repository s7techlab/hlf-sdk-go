package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/s7techlab/hlf-sdk-go/client"
	"github.com/s7techlab/hlf-sdk-go/client/chaincode/system"
	_ "github.com/s7techlab/hlf-sdk-go/crypto/ecdsa"
	"github.com/s7techlab/hlf-sdk-go/identity"
	"github.com/s7techlab/hlf-sdk-go/proto"
)

func main() {

	mspId := "Org1MSP"
	configPath := "./cfg.yaml"

	id, err := identity.FromCertKeyPath(
		mspId,
		// PROVIDE YOUR OWN PATHS
		"../../../../github.com/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/cert.pem",
		"../../../../github.com/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/02a48982a93c9a1fbf7e9702f82d14578aef9662362346ecfe8b3cde50da6799_sk",
	)

	core, err := client.NewCore(id, client.WithConfigYaml(configPath))
	if err != nil {
		log.Fatalln(`unable to initialize core:`, err)
	}

	ctx := context.Background()

	// get chainInfo for all joined channels
	chInfo, err := system.NewCSCC(core, proto.FabricV2).GetChannels(ctx, &empty.Empty{})
	if err != nil {
		log.Fatalln(`failed to fetch channel list:`, err)
	}
	for _, ch := range chInfo.Channels {
		fmt.Printf("Fetching info about channel: %s\n", ch.ChannelId)
		// get blockchain info about channel

		blockchainInfo, err := system.NewQSCC(core).GetChainInfo(ctx, &system.GetChainInfoRequest{ChannelName: ch.ChannelId})
		if err != nil {
			fmt.Println(`Failed to fetch info about channel:`, err)
			continue
		}
		fmt.Printf("Block length: %d, last block: %s, prev block: %s\n", blockchainInfo.Height, base64.StdEncoding.EncodeToString(blockchainInfo.CurrentBlockHash), base64.StdEncoding.EncodeToString(blockchainInfo.PreviousBlockHash))
	}
}
