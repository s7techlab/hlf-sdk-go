## S7 Hyperledger Fabric SDK

[![Coverage Status](https://coveralls.io/repos/github/s7techlab/hlf-sdk-go/badge.svg)](https://coveralls.io/github/s7techlab/hlf-sdk-go)


#### Project structure:

- api - definitions of various cores such as member and operator
- ca - client for Hyperledger Fabric CA
- builder fo channels and chaincodes
- client - for peer and orderer
- crypto - cryptographic implementation
- discovery - discovery service implementation (local and gossip base)
- examples - examples of using current SDK (invoke cli and events client)
    - [event-listener](examples/event-listener) - example of using peer.DeliverService, which shows new blocks
    - [blockchain_info](examples/cc_call/blockchanin_info.go) - example of viewing info about channels and channel's ledger
- identity - member identity implementation
- proto - block parsing

