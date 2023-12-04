package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/service/ccpackage"
	"github.com/s7techlab/hlf-sdk-go/service/ccpackage/fetcher"
	dockerpacker "github.com/s7techlab/hlf-sdk-go/service/ccpackage/packer/docker"
)

func main() {
	spec := &ccpackage.PackageSpec{
		Id: &ccpackage.PackageID{},
	}
	flag.StringVar(&spec.Id.Name, `name`, ``, `chaincode name`)
	flag.StringVar(&spec.Repository, `repo`, ``, `chaincode repo`)
	fabricVersion := flag.String(`fabricVersion`, ``,
		`fabric version (FABRIC_V1, FABRIC_V2, fFABRIC_V2_LIFECYCLE`)
	flag.StringVar(&spec.ChaincodePath, `chaincodePath`, ``, `chaincode path`)
	flag.StringVar(&spec.Id.Version, `version`, ``, `chaincode  version`)
	flag.StringVar(&spec.BinaryPath, `binaryPath`, ``, `binaryPath`)

	flag.Parse()
	if *fabricVersion != `` {
		if enumVersion, ok := ccpackage.FabricVersion_value[*fabricVersion]; ok {
			spec.Id.FabricVersion = ccpackage.FabricVersion(enumVersion)
		} else {
			fmt.Println(`unknown fabric version: `, *fabricVersion)
			os.Exit(1)
		}
	}

	if err := spec.Validate(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var (
		logger, _ = zap.NewDevelopment()
		ctx       = context.Background()
	)

	tar, err := fetcher.Fetch(ctx, spec.Repository, spec.Id.Version, logger)
	if err != nil {
		logger.Fatal(err.Error())
	}

	logger.Info(`package repository source code`, zap.Int(`size`, len(tar)))

	packer := dockerpacker.New(logger)
	pkg, err := packer.PackFromTar(ctx, spec, tar)
	if err != nil {
		logger.Fatal(err.Error())
	}

	fmt.Println(pkg)
}
