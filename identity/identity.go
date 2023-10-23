package identity

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	mspPb "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/crypto"
)

var (
	ErrPEMEncodingFailed = errors.New("pem encoding failed")

	_ msp.SigningIdentity = &SigningIdentity{}
	_ msp.Identity        = &Identity{}
)

type (
	Identity struct {
		certificate *x509.Certificate
		mspId       string
		cryptoSuite crypto.Suite
	}
	SigningIdentity struct {
		*Identity
		privateKey interface{}
	}
)

func New(mspId string, cert *x509.Certificate) *Identity {
	cryptoSuite := crypto.DefaultSuite // todo: ?

	return &Identity{
		mspId:       mspId,
		certificate: cert,
		cryptoSuite: cryptoSuite,
	}
}

func NewSigning(mspId string, cert *x509.Certificate, privateKey interface{}) *SigningIdentity {
	return &SigningIdentity{
		Identity:   New(mspId, cert),
		privateKey: privateKey,
	}
}

func NewSigningFromBytes(mspId string, certRaw []byte, keyRaw []byte) (*SigningIdentity, error) {
	cert, err := Certificate(certRaw)
	if err != nil {
		return nil, fmt.Errorf(`certificate: %w`, err)
	}

	key, err := Key(keyRaw)
	if err != nil {
		return nil, fmt.Errorf(`key: %w`, err)
	}

	return NewSigning(mspId, cert, key), nil
}

func NewSigningFromFile(mspId string, certPath string, keyPath string) (*SigningIdentity, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf(`read certificate from file=%s: %w`, certPath, err)
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf(`read key from file=%s: %w`, keyPath, err)
	}

	return NewSigningFromBytes(mspId, certPEM, keyPEM)
}

func NewSigningFromMSPPath(mspId string, mspPath string) (*SigningIdentity, error) {
	return FirstSigningFromPath(mspId, SignCertsPath(mspPath), KeystorePath(mspPath))
}

func (i *Identity) SigningIdentity(privateKey interface{}) *SigningIdentity {
	return &SigningIdentity{
		Identity:   i,
		privateKey: privateKey,
	}
}

func (i *Identity) GetMSPIdentifier() string {
	return i.mspId
}

func (i *Identity) GetPEM() []byte {
	return PEMEncode(i.certificate.Raw)
}

func (i *Identity) GetCert() *x509.Certificate {
	return i.certificate
}

func (i *Identity) GetIdentifier() *msp.IdentityIdentifier {
	return &msp.IdentityIdentifier{
		Mspid: i.mspId,
		Id:    i.certificate.Subject.CommonName,
	}
}

func (i *Identity) Anonymous() bool {
	return false
}

// ExpiresAt returns date of certificate expiration
func (i *Identity) ExpiresAt() time.Time {
	return i.certificate.NotAfter
}

func (i *Identity) Validate() error {
	// TODO
	return nil
}

func (i *Identity) GetOrganizationalUnits() []*msp.OUIdentifier {
	// TODO
	return nil
}

func (i *Identity) Verify(msg []byte, sig []byte) error {
	return i.cryptoSuite.Verify(i.certificate.PublicKey, msg, sig)
}

func (i *Identity) Serialize() ([]byte, error) {
	pemBytes := PEMEncode(i.certificate.Raw)
	if pemBytes == nil {
		return nil, ErrPEMEncodingFailed
	}

	sId := &mspPb.SerializedIdentity{Mspid: i.mspId, IdBytes: pemBytes}
	idBytes, err := proto.Marshal(sId)
	if err != nil {
		return nil, err
	}

	return idBytes, nil
}

func (i *Identity) SatisfiesPrincipal(principal *mspPb.MSPPrincipal) error {
	panic("implement me")
}

func (s *SigningIdentity) Sign(msg []byte) ([]byte, error) {
	return s.cryptoSuite.Sign(msg, s.privateKey)
}

func (s *SigningIdentity) GetPublicVersion() msp.Identity {
	return s.Identity
}

func PEMEncode(certRaw []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  `CERTIFICATE`,
		Bytes: certRaw,
	})
}
