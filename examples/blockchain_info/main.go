package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	_ "github.com/s7techlab/hlf-sdk-go/crypto/ecdsa"
	_ "github.com/s7techlab/hlf-sdk-go/discovery/local"
	"github.com/s7techlab/hlf-sdk-go/identity"
	"github.com/s7techlab/hlf-sdk-go/member"
)

func main() {
	mspId := os.Getenv(`MSP_ID`)
	if mspId == `` {
		log.Fatalln(`MSP_ID env must be defined`)
	}

	configPath := os.Getenv(`CONFIG_PATH`)
	if configPath == `` {
		log.Fatalln(`CONFIG_PATH env must be defined`)
	}

	certPath := os.Getenv(`CERT_PATH`)
	if certPath == `` {
		log.Fatalln(`CERT_PATH env must be defined`)
	}

	keyPath := os.Getenv(`KEY_PATH`)
	if keyPath == `` {
		log.Fatalln(`KEY_PATH env must be defined`)
	}

	id, err := identity.NewMSPIdentity(mspId, certPath, keyPath)

	if err != nil {
		log.Fatalln(err)
	}

	core, err := member.NewCore(mspId, id, member.WithConfigYaml(configPath))
	if err != nil {
		log.Fatalln(`unable to initialize core:`, err)
	}

	ctx := context.Background()

	// get chainInfo for all joined channels
	if chInfo, err := core.System().CSCC().Channels(ctx); err != nil {
		log.Fatalln(`failed to fetch channel list:`, err)
	} else {
		for _, ch := range chInfo.Channels {
			fmt.Printf("Fetching info about channel: %s\n", ch.ChannelId)
			// get blockchain info about channel
			if blockchainInfo, err := core.System().QSCC().GetChainInfo(ctx, ch.ChannelId); err != nil {
				fmt.Println(`Failed to fetch info about channel:`, err)
			} else {
				fmt.Printf("Block length: %d, last block: %s, prev block: %s\n", blockchainInfo.Height, base64.StdEncoding.EncodeToString(blockchainInfo.CurrentBlockHash), base64.StdEncoding.EncodeToString(blockchainInfo.PreviousBlockHash))
			}
		}
	}
}
