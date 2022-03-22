package identity

import (
	"crypto/x509"
	"fmt"
	"os"

	"github.com/s7techlab/hlf-sdk-go/api"
)

// MSP - contains all parsed identities from msp folder
// Should be used instead of single `api.Identity` which contains ONLY msp identity

type (
	MSPContent struct {
		id string
		// identity from 'signcerts'
		signer api.Identity
		// identities from 'admincerts'
		admins []api.Identity
		// identities from 'users'
		users []api.Identity

		caCerts           []*x509.Certificate
		intermediateCerts []*x509.Certificate
	}

	MSP interface {
		GetMSPIdentifier() string
		Signer() api.Identity
		Admins() []api.Identity
		Users() []api.Identity
		CACerts() []*x509.Certificate
		IntermediateCerts() []*x509.Certificate

		AdminOrSigner() api.Identity
	}

	MspOpts struct {
		mspPath        string
		admincertsPath string
		signcertsPath  string
		keystorePath   string

		loadUsers bool
		userPaths []string

		cacertsPath           string
		intermediatecertsPath string

		loadCertChain     bool
		validateCertChain bool
	}

	MspOpt func(opts *MspOpts)
)

func WithMSPPath(path string) MspOpt {
	return func(opts *MspOpts) {
		opts.mspPath = path
	}
}

func WithCertChain() MspOpt {
	return func(opts *MspOpts) {
		opts.loadCertChain = true
	}
}

func applyDefaultMSPPaths(mspOpts *MspOpts) {
	if mspOpts.admincertsPath == `` {
		mspOpts.admincertsPath = AdmincertsPath(mspOpts.mspPath)
	}

	if mspOpts.signcertsPath == `` {
		mspOpts.signcertsPath = SigncertsPath(mspOpts.mspPath)
	}

	if mspOpts.keystorePath == `` {
		mspOpts.keystorePath = KeystorePath(mspOpts.mspPath)
	}

	if len(mspOpts.userPaths) == 0 && mspOpts.loadUsers {
		mspOpts.userPaths = []string{UsercertsPath(mspOpts.mspPath)}
	}

	if mspOpts.cacertsPath == `` {
		mspOpts.cacertsPath = CacertsPath(mspOpts.mspPath)
	}

	if mspOpts.intermediatecertsPath == `` {
		mspOpts.intermediatecertsPath = IntermediatecertsPath(mspOpts.mspPath)
	}
}

func NewMSP(mspID string, opts ...MspOpt) (*MSPContent, error) {
	var err error
	mspOpts := &MspOpts{}
	for _, opt := range opts {
		opt(mspOpts)
	}

	applyDefaultMSPPaths(mspOpts)

	mspContent := &MSPContent{}

	if mspOpts.admincertsPath != `` {
		mspContent.admins, err = ListFromPath(mspID, mspOpts.admincertsPath, mspOpts.keystorePath)
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf(`read admin identity from=%s: %w`, mspOpts.admincertsPath, err)
		}
	}

	for _, userPath := range mspOpts.userPaths {
		users, err := ListFromPath(mspID, userPath, mspOpts.keystorePath)
		// usePaths set explicit, so if dir is not exists - error occurred
		if err != nil {
			return nil, fmt.Errorf(`read users identity from=%s: %w`, userPath, err)
		}

		mspContent.users = append(mspContent.users, users...)
	}

	if mspOpts.signcertsPath != `` {
		mspContent.signer, err = FirstFromPath(mspID, mspOpts.signcertsPath, mspOpts.keystorePath)
	}

	if mspOpts.loadCertChain {
		mspContent.caCerts, err = CertificatesFromPath(mspOpts.cacertsPath)
		if err != nil {
			return nil, fmt.Errorf(`read cacerts from=%s: %w`, mspOpts.cacertsPath, err)
		}

		mspContent.intermediateCerts, err = CertificatesFromPath(mspOpts.intermediatecertsPath)
		if err != nil {
			return nil, fmt.Errorf(`read intermediatecerts from=%s: %w`, mspOpts.cacertsPath, err)
		}
	}

	if mspOpts.validateCertChain {
		// todo: validate
	}
	return mspContent, nil
}

func (m *MSPContent) GetMSPIdentifier() string {
	return m.id
}

func (m *MSPContent) Signer() api.Identity {
	return m.signer
}

func (m *MSPContent) Admins() []api.Identity {
	return m.admins
}

func (m *MSPContent) Users() []api.Identity {
	return m.users
}

func (m *MSPContent) CACerts() []*x509.Certificate {
	return m.caCerts
}

func (m *MSPContent) IntermediateCerts() []*x509.Certificate {
	return m.intermediateCerts
}

// AdminOrSigner - returns admin identity if exists, in another case return msp.
// installation, fetching  cc list should happen from admin identity
// if there is admin identity, use it. in another case - try with msp identity
func (m *MSPContent) AdminOrSigner() api.Identity {
	if len(m.admins) != 0 {
		return m.admins[0]
	}

	return m.signer
}
