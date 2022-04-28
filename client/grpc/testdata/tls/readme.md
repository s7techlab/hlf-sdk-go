Install https://github.com/cloudflare/cfssl

Issue new certificate
```
cfssl gencert -ca ca.pem -ca-key ca-key.pem csr.json | cfssljson -bare
```