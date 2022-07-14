package main

import (
	"context"
	"log"

	"github.com/atomyze-ru/hlf-sdk-go/client"
	"github.com/atomyze-ru/hlf-sdk-go/identity"
)

func main() {
	mspId := "Org1MSP"
	id, err := identity.FromCertKeyPath(
		mspId,
		// PROVIDE YOUR OWN PATHS
		"../../../../github.com/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/signcerts/cert.pem",
		"../../../../github.com/fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore/02a48982a93c9a1fbf7e9702f82d14578aef9662362346ecfe8b3cde50da6799_sk",
	)
	if err != nil {
		log.Fatalf("connection.invoke: %v", err)
	}

	core, err := client.NewCore(id, client.WithConfigYaml("./cfg.yaml"))
	if err != nil {
		log.Fatalf("create client core: %v", err)
	}
	cc, err := core.Channel("mychannel").Chaincode(context.Background(), "basic")
	if err != nil {
		log.Fatalf("connection.Channel: %v", err)
	}

	res, tx, err := cc.Invoke("UpdateAsset").
		ArgString("asset1", "COLOR", "1337", "OWNER", "228").
		Do(context.Background())
	if err != nil {
		log.Fatalf("connection.invoke: %v", err)
	}

	log.Print("Invoked: ", tx, res)

	res2, err := cc.Query("ReadAsset", "asset1").AsBytes(context.Background())
	if err != nil {
		log.Fatalf("connection.query: %v", err)
	}
	log.Print("Queried: ", string(res2))
}
