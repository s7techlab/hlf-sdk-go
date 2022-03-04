package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/client"
	_ "github.com/s7techlab/hlf-sdk-go/crypto/ecdsa"
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

	core, err := client.NewCore(id, client.WithConfigYaml(configPath), client.WithLogger(l))
	if err != nil {
		log.Fatalln(`unable to initialize core:`, err)
	}

	cc, err := core.Channel(channel).Chaincode(context.Background(), chaincode)
	if err != nil {
		log.Fatalln(`unable to initialize channel:`, err)
	}

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func(idx int, wg *sync.WaitGroup) {
			defer wg.Done()
			sub, err := cc.Subscribe(context.Background())
			if err != nil {
				log.Printf("Failed to process rouitine %d: %s", idx, err)
				return
			}
			defer func() { _ = sub.Close() }()

			for {
				select {
				case ev := <-sub.Events():
					b, _ := json.MarshalIndent(ev, ` `, "\t")
					fmt.Printf("Routine %d, received event:\n %v\n", idx, string(b))
				case err := <-sub.Errors():
					log.Println(`error occurred:`, err)
					return
				case <-time.After(time.Duration(idx) * time.Second):
					fmt.Printf("Routine %d is closing\n", idx)
					return
				}
			}

		}(i, &wg)
	}

	wg.Wait()

}
