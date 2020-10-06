# Install and instantiate cli

Tool for installing and instantiating chaincode

Required flags:
- mspId - identifier of MSP
- mspPath - path to `msp` directory
- configPath - path to SDK config ([example](../caclient/config.yaml))
- channel - channel name for instantiated chaincode
- cc - chaincode name
- ccPath - path to chaincode realtively `$GOPATH`
- ccVersion - chaincode version
- ccPolicy - instantiation policy
- ccArgs - chaincode instantiation arguments
- ccTransient - chaincode transient arguments
