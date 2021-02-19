package main

import (
	"log"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func main() {
	if err := shim.Start(&example_cc{}); err != nil {
		log.Fatal(err)
	}
}

type example_cc struct {
}

func (*example_cc) Init(stub shim.ChaincodeStubInterface) peer.Response {
	t, err := stub.GetTransient()
	if err != nil {
		return shim.Error(err.Error())
	}

	log.Println(t)

	if err = stub.PutState(`key`, t[`key`]); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (*example_cc) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	t, err := stub.GetTransient()
	if err != nil {
		return shim.Error(err.Error())
	}

	log.Println(t)

	if err = stub.PutState(`key`, t[`key`]); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}
