package main

import (
	"context"
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

	channel := os.Getenv(`CHANNEL`)
	if channel == `` {
		log.Fatalln(`CHANNEL env must be defined`)
	}

	chaincode := os.Getenv(`CHAINCODE`)
	if chaincode == `` {
		log.Fatalln(`CHAINCODE env must be defined`)
	}

	id, err := identity.NewMSPIdentity(mspId, certPath, keyPath)

	if err != nil {
		log.Fatalln(err)
	}

	core, err := member.NewCore(mspId, id, member.WithConfigYaml(configPath))
	if err != nil {
		log.Fatalln(`unable to initialize core:`, err)
	}

	cc := core.Channel(channel).Chaincode(chaincode)
	sub := cc.Subscribe(context.Background())
	if evChan, err := sub.Events(); err != nil {
		log.Fatalln(`failed to subscribe on events:`, err)
	} else {
		fmt.Printf("Waiting for events on chaincode `%s` from channel `%s`...\n", chaincode, channel)
		for ev := range evChan {
			fmt.Printf("Received event: %v\n", *ev)
		}
	}
}
