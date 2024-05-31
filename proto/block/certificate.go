package block

import (
	"crypto/sha256"
	"encoding/pem"
	"fmt"
)

func NewCertificate(cert []byte, t CertType, mspID, mspName string) (*Certificate, error) {
	b, _ := pem.Decode(cert)
	if b == nil {
		return &Certificate{}, fmt.Errorf("decode %s cert of %s", t, mspID)
	}

	c := &Certificate{
		Data:    cert,
		MspId:   mspID,
		Type:    t,
		MspName: mspName,
	}
	c.setCertificateSHA256(b)

	return c, nil
}

func (x *Certificate) setCertificateSHA256(b *pem.Block) {
	f := CalcCertificateSHA256(b)
	x.Fingerprint = f[:]
}

func CalcCertificateSHA256(b *pem.Block) []byte {
	f := sha256.Sum256(b.Bytes)
	return f[:]
}
