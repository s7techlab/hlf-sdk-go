# Generate test CA keypair

cfssl genkey -initca csr.json | cfssljson -bare ca
