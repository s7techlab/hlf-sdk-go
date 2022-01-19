package identity

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/golang/protobuf/proto"
	mspPb "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric/msp"
	_ "github.com/hyperledger/fabric/msp"
	"github.com/pkg/errors"

	"github.com/s7techlab/hlf-sdk-go/v2/api"
	"github.com/s7techlab/hlf-sdk-go/v2/util"
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
	id := s.signingIdentity
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
		return nil, errors.Wrap(api.ErrInvalidPEMStructure, `failed to decode certificate`)
	}

	keyPEM, _ := pem.Decode(keyBytes)
	if keyPEM == nil {
		return nil, errors.Wrap(api.ErrInvalidPEMStructure, `failed to decode private key`)
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

	certBytes, keyBytes, err := util.LoadKeyPairFromMSP(mspPath)
	if err != nil {
		return nil, err
	}

	return NewMSPIdentityBytes(mspId, certBytes, keyBytes)
}

/* */

// MSPIdentities - contains all parsed identities from msp folder
// Should be used instead of single `api.Identity` which contains ONLY msp identity
type MSPIdentities struct {
	// identity from 'signcerts'
	MSP api.Identity
	// identities from 'admincerts'
	Admins []api.Identity
	// identities from 'users'
	Users []api.Identity
}

// TODO some handy methods?

// NewMSPIndentitiesFromPath - parse all certificates(msp,admins,users) from MSP folder.
// Came to replace legacy `util.LoadKeyPairFromMSP` method
func NewMSPIndentitiesFromPath(mspID string, mspPath string) (*MSPIdentities, error) {
	const (
		admincertsPath = "admincerts"
		signcertsPath  = "signcerts"
		userscertsPath = "user"
	)

	mspIdentities := &MSPIdentities{
		Admins: make([]api.Identity, 0),
		MSP:    nil,
		Users:  make([]api.Identity, 0),
	}

	// admin certs
	adminDir := filepath.Join(mspPath, admincertsPath)
	_, err := os.Stat(adminDir)
	if !os.IsNotExist(err) {
		adminCerts, err := util.ReadAllFilesFromDir(adminDir)
		if err != nil {
			return nil, err
		}

		for _, adminCertBytes := range adminCerts {
			cert, key, err := util.LoadKeypairByCert(mspPath, adminCertBytes)
			if err != nil {
				return nil, err
			}

			idnt, _ := NewMSPIdentityRaw(mspID, cert, key)
			mspIdentities.Admins = append(mspIdentities.Admins, idnt)
		}
	}

	// user certs
	userDir := filepath.Join(mspPath, userscertsPath)
	_, err = os.Stat(userDir)
	if !os.IsNotExist(err) {
		userCerts, err := util.ReadAllFilesFromDir(userDir)
		if err != nil {
			return nil, err
		}

		for _, userCertBytes := range userCerts {
			cert, key, err := util.LoadKeypairByCert(mspPath, userCertBytes)
			if err != nil {
				return nil, err
			}

			idnt, _ := NewMSPIdentityRaw(mspID, cert, key)
			mspIdentities.Users = append(mspIdentities.Users, idnt)
		}
	}

	// signcert
	signCertsDir := filepath.Join(mspPath, signcertsPath)
	signCerts, err := util.ReadAllFilesFromDir(signCertsDir)
	if err != nil {
		return nil, err
	}
	if len(signCerts) == 0 {
		return nil, errors.Wrap(err, `'signcerts' folder is emprty`)
	}

	cert, key, err := util.LoadKeypairByCert(mspPath, signCerts[0])
	if err != nil {
		return nil, err
	}

	idnt, _ := NewMSPIdentityRaw(mspID, cert, key)
	mspIdentities.MSP = idnt

	return mspIdentities, nil
}
