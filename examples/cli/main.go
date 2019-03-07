package main

import (
	"context"
	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client"
	_ "github.com/s7techlab/hlf-sdk-go/crypto/ecdsa"
	_ "github.com/s7techlab/hlf-sdk-go/discovery/local"
	"github.com/s7techlab/hlf-sdk-go/identity"
	"go.uber.org/zap"
	"log"
	"os"
)

var ctx = context.Background()

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

	id, err := identity.NewMSPIdentityFromPath(mspId, mspPath)

	if err != nil {
		log.Fatalln(`Failed to load identity:`, err)
	}

	l, _ := zap.NewDevelopment()

	core, err := client.NewCore(mspId, id, client.WithConfigYaml(configPath), client.WithLogger(l))
	if err != nil {
		log.Fatalln(`unable to initialize core:`, err)
	}

	//if err = core.Chaincode(`example`).Install(ctx, `github.com/s7techlab/hlf-sdk-go/samples/example_cc`, `0.1`); err != nil {
	//	log.Fatalln(err)
	//}

	if err = core.Chaincode(`example`).Instantiate(
		ctx,
		`public`,
		`github.com/s7techlab/hlf-sdk-go/samples/example_cc`,
		`0.1`,
		`AND ("OPERATORMSP.admin")`,
		[][]byte{},
		api.TransArgs{`key`: []byte(`value`)},
	); err != nil {
		log.Fatalln(err)
	}
}
