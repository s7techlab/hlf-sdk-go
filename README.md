## S7 Hyperledger Fabric SDK

[![Coverage Status](https://coveralls.io/repos/github/s7techlab/hlf-sdk-go/badge.svg)](https://coveralls.io/github/s7techlab/hlf-sdk-go)

Alpha version, **use at your own risk!**

#### Project structure:

- api - definitions of various cores such as member and operator
- crypto - cryptographic implementation
- discovery - discovery service implementation (local only)
- examples - examples of using current SDK (invoke cli and events client)
    - [event-listener](examples/event-listener) - example of using peer.DeliverService, which shows new blocks
    - [blockchain_info](examples/blockchain_info) - example of viewing info about channels and channel's ledger
- identity - member identity implementation
- member - member core implementation

#### Thanks
- [GoHFC](https://github.com/CognitionFoundry/gohfc) - for basic ideas and examples