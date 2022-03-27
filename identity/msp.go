package identity

import (
	"crypto/x509"
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/s7techlab/hlf-sdk-go/api"
)

// MSP - contains all parsed identities from msp folder
// Should be used instead of single `api.Identity` which contains ONLY msp identity

type (
	MSPConfig struct {
		id string
		// identity from 'signcerts'
		signer api.Identity
		// identities from 'admincerts'
		admins []api.Identity
		// identities from 'users'
		users []api.Identity

		caCerts           []*x509.Certificate
		intermediateCerts []*x509.Certificate

		ouConfig *OUConfig
	}

	MSP interface {
		GetMSPIdentifier() string
		Signer() api.Identity
		Admins() []api.Identity
		Users() []api.Identity
		CACerts() []*x509.Certificate
		IntermediateCerts() []*x509.Certificate

		AdminOrSigner() api.Identity

		OUConfig() *OUConfig
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

		loadOUConfig bool

		logger *zap.Logger
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

func WithOUConfig() MspOpt {
	return func(opts *MspOpts) {
		opts.loadOUConfig = true
	}
}

func applyDefaultMSPPaths(mspOpts *MspOpts) {
	if mspOpts.mspPath != `` {
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
}

func NewMSP(mspID string, opts ...MspOpt) (*MSPConfig, error) {
	var err error
	mspOpts := &MspOpts{}
	for _, opt := range opts {
		opt(mspOpts)
	}

	logger := zap.NewNop()
	if mspOpts.logger != nil {
		logger = mspOpts.logger
	}

	applyDefaultMSPPaths(mspOpts)

	logger.Debug(`load msp`, zap.Reflect(`config`, mspOpts))

	mspConfig := &MSPConfig{}

	if mspOpts.admincertsPath != `` {
		mspConfig.admins, err = ListFromPath(mspID, mspOpts.admincertsPath, mspOpts.keystorePath)
		if err != nil {
			logger.Debug(`load admin identities`, zap.Error(err))
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf(`read admin identity from=%s: %w`, mspOpts.admincertsPath, err)
			}
		}

		logger.Debug(`admin identities loaded`, zap.Int(`num`, len(mspConfig.admins)))
	}

	if len(mspOpts.userPaths) > 0 {
		for _, userPath := range mspOpts.userPaths {
			users, err := ListFromPath(mspID, userPath, mspOpts.keystorePath)
			// usePaths set explicit, so if dir is not exists - error occurred
			if err != nil {
				return nil, fmt.Errorf(`read users identity from=%s: %w`, userPath, err)
			}

			mspConfig.users = append(mspConfig.users, users...)
		}

		logger.Debug(`user identities loaded`, zap.Int(`num`, len(mspConfig.users)))
	}

	if mspOpts.signcertsPath != `` {
		mspConfig.signer, err = FirstFromPath(mspID, mspOpts.signcertsPath, mspOpts.keystorePath)
		if err != nil {
			return nil, fmt.Errorf(`read signer identity from=%s: %w`, mspOpts.signcertsPath, err)
		}
	}

	if mspOpts.loadCertChain {
		mspConfig.caCerts, err = CertificatesFromPath(mspOpts.cacertsPath)
		if err != nil {
			if os.IsNotExist(err) {
				logger.Debug(`cacerts path not found`, zap.String(`path`, mspOpts.cacertsPath))
			} else {
				return nil, fmt.Errorf(`read cacerts from=%s: %w`, mspOpts.cacertsPath, err)
			}
		}

		logger.Debug(`CA certs loaded`, zap.Int(`num`, len(mspConfig.caCerts)))

		mspConfig.intermediateCerts, err = CertificatesFromPath(mspOpts.intermediatecertsPath)
		if err != nil {
			if os.IsNotExist(err) {
				logger.Debug(`intermediatecerts path not found`, zap.String(`path`, mspOpts.intermediatecertsPath))
			} else {
				return nil, fmt.Errorf(`read intermediatecerts from=%s: %w`, mspOpts.cacertsPath, err)
			}
		}

		logger.Debug(`intermediate certs loaded`, zap.Int(`num`, len(mspConfig.caCerts)))
	}

	if mspOpts.validateCertChain {
		// todo: validate
	}

	if mspOpts.mspPath != `` && mspOpts.loadOUConfig {
		if mspConfig.ouConfig, err = ReadNodeOUConfig(mspOpts.mspPath); err != nil {
			return nil, err
		}
	}
	return mspConfig, nil
}

func (m *MSPConfig) GetMSPIdentifier() string {
	return m.id
}

func (m *MSPConfig) Signer() api.Identity {
	return m.signer
}

func (m *MSPConfig) Admins() []api.Identity {
	return m.admins
}

func (m *MSPConfig) Users() []api.Identity {
	return m.users
}

func (m *MSPConfig) CACerts() []*x509.Certificate {
	return m.caCerts
}

func (m *MSPConfig) IntermediateCerts() []*x509.Certificate {
	return m.intermediateCerts
}

// AdminOrSigner - returns admin identity if exists, in another case return msp.
// installation, fetching  cc list should happen from admin identity
// if there is admin identity, use it. in another case - try with msp identity
func (m *MSPConfig) AdminOrSigner() api.Identity {
	if len(m.admins) != 0 {
		return m.admins[0]
	}

	return m.signer
}

func (m *MSPConfig) OUConfig() *OUConfig {
	return m.ouConfig
}
