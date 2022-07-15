### Blockchain Info

Example allows to get information about joined channels and channel's ledger.

#### Required environment variables
- **MSP_ID** - MSP identifier
- **CONFIG_PATH** - path to configuration file of SDK _(eg. config.yaml)_
- **CERT_PATH** - path to certificate file of identity _(eg. signcerts/cert.pem)_
- **KEY_PATH** - path to private key of identity _(eg. keystore/your_pk)_

#### Usage
Run with environment variables described above:
```bash
go run main.go
```

Example output:
```bash
Fetching info about channel: channel1
Block length: 1, last block: hSUcm1BXeBFUDkYoktP2snFh2sdMtrO7Wn4e381Dq4Y=, prev block: 
Fetching info about channel: channel2
Block length: 2, last block: NJdmpdTjGZG3NDfJKWD1qUsNWs2DI/eQnCRBRrioVKo=, prev block: 6vm07LCMvpIjJ9OZ54WJZfZNKZ21w2KbSkEVExjvfN4=
Fetching info about channel: channel3
Block length: 1, last block: eX8As3kxjTbep4QMCGuzj/wrZQQhNlMNTS+zGOBd+lk=, prev block:
```