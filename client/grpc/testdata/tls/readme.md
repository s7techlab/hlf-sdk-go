# Install https://github.com/cloudflare/cfssl

## Issue CA cert

```
cd ./ca && cfssl genkey -initca csr.json | cfssljson -bare ca
```

## Issue client certificate

```
(cd ./client && cfssl gencert -ca ../ca/ca.pem -ca-key ../ca/ca-key.pem -config ../ca/cfssl.json -profile client csr.json | cfssljson -bare)
```

## Issue server certificate

```
(cd ./server && cfssl gencert -ca ../ca/ca.pem -ca-key ../ca/ca-key.pem -config ../ca/cfssl.json -profile server csr.json | cfssljson -bare)
```