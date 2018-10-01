package identity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"path"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/msp"
	_ "github.com/hyperledger/fabric/msp"
	mspPb "github.com/hyperledger/fabric/protos/msp"
	"github.com/pkg/errors"
	"github.com/s7techlab/hlf-sdk-go/api"
)

const (
	signCertPath = `signcerts`
	keyStorePath = `keystore`
)

var (
	errInvalidPEMStructure = errors.New(`invalid PEM structure`)
)

type mspIdentity struct {
	signingIdentity *mspSigningIdentity
}

type mspSigningIdentity struct {
	privateKey  interface{}
	publicKey   interface{}
	certificate *x509.Certificate
	mspId       string
	cryptoSuite api.CryptoSuite
}

func (s *mspIdentity) GetSigningIdentity(cs api.CryptoSuite) msp.SigningIdentity {
	id := new(mspSigningIdentity)
	id = s.signingIdentity
	id.cryptoSuite = cs
	return id
}

func (s *mspSigningIdentity) Anonymous() bool {
	return false
}

// ExpiresAt returns date of certificate expiration
func (s *mspSigningIdentity) ExpiresAt() time.Time {
	return s.certificate.NotAfter
}

func (s *mspSigningIdentity) GetIdentifier() *msp.IdentityIdentifier {
	return &msp.IdentityIdentifier{
		Mspid: s.mspId,
		Id:    s.certificate.Subject.CommonName,
	}
}

// GetMSPIdentifier returns current MspID of identity
func (s *mspSigningIdentity) GetMSPIdentifier() string {
	return s.mspId
}

func (s *mspSigningIdentity) Validate() error {
	// TODO
	return nil
}

func (s *mspSigningIdentity) GetOrganizationalUnits() []*msp.OUIdentifier {
	// TODO
	return nil
}

func (s *mspSigningIdentity) Verify(msg []byte, sig []byte) error {
	return s.cryptoSuite.Verify(s.publicKey, msg, sig)
}

func (s *mspSigningIdentity) Serialize() ([]byte, error) {
	pb := &pem.Block{Bytes: s.certificate.Raw, Type: "CERTIFICATE"}
	pemBytes := pem.EncodeToMemory(pb)
	if pemBytes == nil {
		return nil, errors.New("encoding of identity failed")
	}

	sId := &mspPb.SerializedIdentity{Mspid: s.mspId, IdBytes: pemBytes}
	idBytes, err := proto.Marshal(sId)
	if err != nil {
		return nil, err
	}

	return idBytes, nil
}

func (s *mspSigningIdentity) SatisfiesPrincipal(principal *mspPb.MSPPrincipal) error {
	panic("implement me")
}

func (s *mspSigningIdentity) Sign(msg []byte) ([]byte, error) {
	return s.cryptoSuite.Sign(msg, s.privateKey)
}

func (s *mspSigningIdentity) GetPublicVersion() msp.Identity {
	return nil
}

func NewMSPIdentity(mspId string, certPath string, keyPath string) (api.Identity, error) {
	certPEMBytes, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, errors.Wrap(err, `failed to open certificate`)
	}

	keyPEMBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, errors.Wrap(err, `failed to open private key`)
	}

	return NewMSPIdentityBytes(mspId, certPEMBytes, keyPEMBytes)
}

func NewMSPIdentityBytes(mspId string, certBytes []byte, keyBytes []byte) (api.Identity, error) {
	certPEM, _ := pem.Decode(certBytes)
	if certPEM == nil {
		return nil, errInvalidPEMStructure
	}

	keyPEM, _ := pem.Decode(keyBytes)
	if keyPEM == nil {
		return nil, errInvalidPEMStructure
	}

	cert, err := x509.ParseCertificate(certPEM.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, `failed to parse x509 certificate`)
	}

	key, err := x509.ParsePKCS8PrivateKey(keyPEM.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, `failed to parse private key`)
	}

	return NewMSPIdentityRaw(mspId, cert, key)
}

func NewMSPIdentityRaw(mspId string, cert *x509.Certificate, privateKey interface{}) (api.Identity, error) {

	signingIdentity := &mspSigningIdentity{mspId: mspId, privateKey: privateKey, certificate: cert, publicKey: cert.PublicKey}

	return &mspIdentity{signingIdentity: signingIdentity}, nil
}

func NewEnrollIdentity(privateKey interface{}) (api.Identity, error) {
	identity := &mspSigningIdentity{privateKey: privateKey}
	return &mspIdentity{signingIdentity: identity}, nil
}

func NewMSPIdentityFromPath(mspId string, mspPath string) (api.Identity, error) {
	_, err := ioutil.ReadDir(mspPath)
	if err != nil {
		return nil, errors.Wrap(err, `failed to read path`)
	}

	// check signcerts/cert.pem
	certBytes, err := ioutil.ReadFile(path.Join(mspPath, signCertPath, `cert.pem`))
	if err != nil {
		return nil, errors.Wrap(err, `failed to read certificate`)
	}

	certBlock, _ := pem.Decode(certBytes)
	if certBlock == nil {
		return nil, errInvalidPEMStructure
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, `failed to parse certificate`)
	}

	var pKeyFileName string

	switch pubKey := cert.PublicKey.(type) {
	case *ecdsa.PublicKey:
		h := sha256.New()
		h.Write(elliptic.Marshal(pubKey.Curve, pubKey.X, pubKey.Y))
		pKeyFileName = fmt.Sprintf("%x_sk", h.Sum(nil))
	default:
		return nil, errors.Errorf("unknown key format %s, ECDSA expected", cert.PublicKeyAlgorithm)
	}

	keyBytes, err := ioutil.ReadFile(path.Join(mspPath, keyStorePath, pKeyFileName))
	if err != nil {
		return nil, errors.Wrap(err, `failed to ready private key file`)
	}

	return NewMSPIdentityBytes(mspId, certBytes, keyBytes)

}
