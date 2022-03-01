package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/s7techlab/hlf-sdk-go/v2/client"
	_ "github.com/s7techlab/hlf-sdk-go/v2/crypto/ecdsa"
	"github.com/s7techlab/hlf-sdk-go/v2/identity"
)

func main() {
	// mspId := os.Getenv(`MSP_ID`)
	// if mspId == `` {
	// 	log.Fatalln(`MSP_ID env must be defined`)
	// }

	// configPath := os.Getenv(`CONFIG_PATH`)
	// if configPath == `` {
	// 	log.Fatalln(`CONFIG_PATH env must be defined`)
	// }

	// identityPath := os.Getenv(`IDENTITY_PATH`)
	// if identityPath == `` {
	// 	log.Fatalln(`KEY_PATH env must be defined`)
	// }
	mspId := "Org1MSP"
	configPath := "./cfg.yaml"

	id, err := identity.NewMSPIdentity(
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
	chInfo, err := core.System().CSCC().GetChannels(ctx)
	if err != nil {
		log.Fatalln(`failed to fetch channel list:`, err)
	}
	for _, ch := range chInfo.Channels {
		fmt.Printf("Fetching info about channel: %s\n", ch.ChannelId)
		// get blockchain info about channel
		blockchainInfo, err := core.System().QSCC().GetChainInfo(ctx, ch.ChannelId)
		if err != nil {
			fmt.Println(`Failed to fetch info about channel:`, err)
			continue
		}
		fmt.Printf("Block length: %d, last block: %s, prev block: %s\n", blockchainInfo.Height, base64.StdEncoding.EncodeToString(blockchainInfo.CurrentBlockHash), base64.StdEncoding.EncodeToString(blockchainInfo.PreviousBlockHash))
	}
}
