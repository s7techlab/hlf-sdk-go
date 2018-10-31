package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/client"

	_ "github.com/s7techlab/hlf-sdk-go/crypto/ecdsa"
	_ "github.com/s7techlab/hlf-sdk-go/discovery/local"
	"github.com/s7techlab/hlf-sdk-go/identity"
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

	mspPath := os.Getenv(`MSP_PATH`)
	if mspPath == `` {
		log.Fatalln(`MSP_PATH env must be defined`)
	}

	channel := os.Getenv(`CHANNEL`)
	if channel == `` {
		log.Fatalln(`CHANNEL env must be defined`)
	}

	chaincode := os.Getenv(`CHAINCODE`)
	if chaincode == `` {
		log.Fatalln(`CHAINCODE env must be defined`)
	}

	id, err := identity.NewMSPIdentityFromPath(mspId, mspPath)

	if err != nil {
		log.Fatalln(`Failed to load identity:`, err)
	}

	l, _ := zap.NewProduction()

	core, err := client.NewCore(mspId, id, client.WithConfigYaml(configPath), client.WithLogger(l))
	if err != nil {
		log.Fatalln(`unable to initialize core:`, err)
	}

	cc := core.Channel(channel).Chaincode(chaincode)
	sub, err := cc.Subscribe(context.Background())
	if err != nil {
		log.Fatalln(`failed to get sub:`, err)
	}
	fmt.Printf("Waiting for events on chaincode `%s` from channel `%s`...\n", chaincode, channel)
	for ev := range sub.Events() {
		fmt.Printf("Received event: %v\n", *ev)
	}
}
