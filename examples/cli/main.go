package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"

	"github.com/hyperledger/fabric/common/util"
	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/api"
	"github.com/s7techlab/hlf-sdk-go/client"
	_ "github.com/s7techlab/hlf-sdk-go/crypto/ecdsa"
	"github.com/s7techlab/hlf-sdk-go/identity"
)

var ctx = context.Background()

var (
	mspId      = flag.String(`mspId`, ``, `MspId`)
	mspPath    = flag.String(`mspPath`, ``, `path to admin certificate`)
	configPath = flag.String(`configPath`, ``, `path to configuration file`)

	channel     = flag.String(`channel`, ``, `channel name`)
	cc          = flag.String(`cc`, ``, `chaincode name`)
	ccPath      = flag.String(`ccPath`, ``, `chaincode path`)
	ccVersion   = flag.String(`ccVersion`, ``, `chaincode version`)
	ccPolicy    = flag.String(`ccPolicy`, ``, `chaincode endorsement policy`)
	ccArgs      = flag.String(`ccArgs`, ``, `chaincode instantiation arguments`)
	ccTransient = flag.String(`ccTransient`, ``, `chaincode transient arguments`)
)

func main() {
	id, err := identity.NewMSPIdentityFromPath(*mspId, *mspPath)

	if err != nil {
		log.Fatalln(`Failed to load identity:`, err)
	}

	l, _ := zap.NewDevelopment()

	core, err := client.NewCore(id, client.WithConfigYaml(*configPath), client.WithLogger(l))
	if err != nil {
		log.Fatalln(`unable to initialize core:`, err)
	}

	if err = core.Chaincode(*cc).Install(ctx, *ccPath, *ccVersion); err != nil {
		log.Fatalln(err)
	}

	if err = core.Chaincode(*cc).Instantiate(
		ctx,
		*channel,
		*ccPath,
		*ccVersion,
		*ccPolicy,
		util.ToChaincodeArgs(*ccArgs),
		prepareTransArgs(*ccTransient),
	); err != nil {
		log.Fatalln(err)
	}

	log.Println(`successfully initiated`)
}

func prepareTransArgs(args string) api.TransArgs {
	var t map[string]json.RawMessage
	var err error
	if err = json.Unmarshal([]byte(args), &t); err != nil {
		panic(err)
	}

	tt := api.TransArgs{}

	for k, v := range t {
		if tt[k], err = v.MarshalJSON(); err != nil {
			panic(err)
		}
	}
	return tt
}

func init() {
	flag.Parse()
}
