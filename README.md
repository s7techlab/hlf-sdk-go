## Hyperledger Fabric SDK

[![Coverage Status](https://coveralls.io/repos/github/atomyze-ru/hlf-sdk-go/badge.svg)](https://coveralls.io/github/atomyze-ru/hlf-sdk-go)

Code example with gossip service discovery available at: `examples/cc_call`
#### Project structure:

- api - interface definitions
- ca - sdk for Hyperledger Fabric CA
- client - sdk for Hyperledger Fabric Network
- crypto - cryptographic implementation
- discovery - discovery service implementation
- examples - examples of using current SDK (invoke cli and events client)
    - [event-listener](examples/event-listener) - example of using peer.DeliverService, which shows new blocks
    - [blockchain_info](examples/channel_info/blockchain_info.go) - example of viewing info about channels and channel's ledger
- identity - identity implementation
- proto - Hyperledger fabric protobuf messages creating and parsing
