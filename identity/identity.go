package identity

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/golang/protobuf/proto"
	mspPb "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric/msp"
	_ "github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"

	"github.com/atomyze-ru/hlf-sdk-go/api"
)

type (
	identity struct {
		privateKey  interface{}
		publicKey   interface{}
		certificate *x509.Certificate
		mspId       string
	}
	signingIdentity struct {
		identity    *identity
		cryptoSuite api.CryptoSuite
	}
)

func New(mspId string, cert *x509.Certificate, privateKey interface{}) *identity {
	return &identity{
		mspId:       mspId,
		privateKey:  privateKey,
		certificate: cert,
		publicKey:   cert.PublicKey,
	}
}

func FromBytesWithoutSigning(mspId string, certRaw []byte) (*identity, error) {
	cert, err := Certificate(certRaw)
	if err != nil {
		return nil, fmt.Errorf(`certificate: %w`, err)
	}

	return New(mspId, cert, nil), nil
}

func FromBytes(mspId string, certRaw []byte, keyRaw []byte) (*identity, error) {
	cert, err := Certificate(certRaw)
	if err != nil {
		return nil, fmt.Errorf(`certificate: %w`, err)
	}

	key, err := Key(keyRaw)
	if err != nil {
		return nil, fmt.Errorf(`key: %w`, err)
	}

	return New(mspId, cert, key), nil
}

func FromCertKeyPath(mspId string, certPath string, keyPath string) (api.Identity, error) {
	certPEM, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf(`read certificate from file=%s: %w`, certPath, err)
	}

	keyPEM, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf(`read key from file=%s: %w`, keyPath, err)
	}

	return FromBytes(mspId, certPEM, keyPEM)
}

func SignerFromMSPPath(mspId string, mspPath string) (*identity, error) {
	return FirstFromPath(mspId, SignCertsPath(mspPath), KeystorePath(mspPath))
}

func (i *identity) GetSigningIdentity(cs api.CryptoSuite) msp.SigningIdentity {
	return &signingIdentity{
		identity:    i,
		cryptoSuite: cs,
	}
}

func (i *identity) GetMSPIdentifier() string {
	return i.mspId
}

func PEMEncode(certRaw []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{
		Type:  `CERTIFICATE`,
		Bytes: certRaw,
	})
}

func (i *identity) GetPEM() []byte {
	return PEMEncode(i.certificate.Raw)
}

func (i *identity) GetCert() *x509.Certificate {
	return i.certificate
}

func (i *identity) GetIdentifier() *msp.IdentityIdentifier {
	return &msp.IdentityIdentifier{
		Mspid: i.mspId,
		Id:    i.certificate.Subject.CommonName,
	}
}

func (s *signingIdentity) Anonymous() bool {
	return false
}

// ExpiresAt returns date of certificate expiration
func (s *signingIdentity) ExpiresAt() time.Time {
	return s.identity.certificate.NotAfter
}

func (s *signingIdentity) GetIdentifier() *msp.IdentityIdentifier {
	return s.identity.GetIdentifier()
}

// GetMSPIdentifier returns current MspID of identity
func (s *signingIdentity) GetMSPIdentifier() string {
	return s.identity.GetMSPIdentifier()
}

func (s *signingIdentity) Validate() error {
	// TODO
	return nil
}

func (s *signingIdentity) GetOrganizationalUnits() []*msp.OUIdentifier {
	// TODO
	return nil
}

func (s *signingIdentity) Verify(msg []byte, sig []byte) error {
	return s.cryptoSuite.Verify(s.identity.publicKey, msg, sig)
}

func (s *signingIdentity) Serialize() ([]byte, error) {
	pb := &pem.Block{Bytes: s.identity.certificate.Raw, Type: "CERTIFICATE"}
	pemBytes := pem.EncodeToMemory(pb)
	if pemBytes == nil {
		return nil, errors.New("encoding of identity failed")
	}

	sId := &mspPb.SerializedIdentity{Mspid: s.identity.mspId, IdBytes: pemBytes}
	idBytes, err := proto.Marshal(sId)
	if err != nil {
		return nil, err
	}

	return idBytes, nil
}

func (s *signingIdentity) SatisfiesPrincipal(principal *mspPb.MSPPrincipal) error {
	panic("implement me")
}

func (s *signingIdentity) Sign(msg []byte) ([]byte, error) {
	return s.cryptoSuite.Sign(msg, s.identity.privateKey)
}

func (s *signingIdentity) GetPublicVersion() msp.Identity {
	return nil
}
