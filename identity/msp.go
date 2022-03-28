package identity

import (
	"crypto/x509"
	"fmt"
	"os"

	"github.com/hyperledger/fabric-protos-go/msp"
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

		caCerts              []*x509.Certificate
		intermediateCerts    []*x509.Certificate
		tlsCACerts           []*x509.Certificate
		tlsIntermediateCerts []*x509.Certificate

		ouConfig *OUConfig
	}

	MSPConfigSerialized struct {
		// Certs path to cert file (for example admincerts/cert.pem) => cert content
		Certs MSPCerts
		OU    *OUConfigSerialized
	}

	MSPCerts map[string][]byte

	MSP interface {
		GetMSPIdentifier() string
		Signer() api.Identity
		Admins() []api.Identity
		Users() []api.Identity
		CACerts() []*x509.Certificate
		IntermediateCerts() []*x509.Certificate
		TLSCACerts() []*x509.Certificate
		TLSIntermediateCerts() []*x509.Certificate
		AdminOrSigner() api.Identity

		OUConfig() *OUConfig
		FabricMSPConfig() *msp.FabricMSPConfig
	}

	MspOpts struct {
		mspPath string

		adminCertsPath           string
		signCertsPath            string
		keystorePath             string
		loadUsers                bool
		userPaths                []string
		caCertsPath              string
		intermediateCertsPath    string
		tlsCaCertsPath           string
		tlsIntermediateCertsPath string

		loadCertChain     bool
		validateCertChain bool

		loadOUConfig bool

		logger *zap.Logger
	}

	MspOpt func(opts *MspOpts)
)

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

	if mspOpts.adminCertsPath == `` {
		mspOpts.adminCertsPath = AdminCertsPath(mspOpts.mspPath)
	}

	if mspOpts.signCertsPath == `` {
		mspOpts.signCertsPath = SignCertsPath(mspOpts.mspPath)
	}

	if mspOpts.keystorePath == `` {
		mspOpts.keystorePath = KeystorePath(mspOpts.mspPath)
	}

	if len(mspOpts.userPaths) == 0 && mspOpts.loadUsers {
		mspOpts.userPaths = []string{UsercertsPath(mspOpts.mspPath)}
	}

	if mspOpts.caCertsPath == `` {
		mspOpts.caCertsPath = CACertsPath(mspOpts.mspPath)
	}

	if mspOpts.intermediateCertsPath == `` {
		mspOpts.intermediateCertsPath = IntermediateCertsPath(mspOpts.mspPath)
	}

	if mspOpts.tlsCaCertsPath == `` {
		mspOpts.tlsCaCertsPath = TLSCACertsPath(mspOpts.mspPath)
	}

	if mspOpts.tlsIntermediateCertsPath == `` {
		mspOpts.tlsIntermediateCertsPath = TLSIntermediateCertsPath(mspOpts.mspPath)
	}

}

// MSPFromConfig loads msp config from msp.FabricMSPConfig)
func MSPFromConfig(config *msp.FabricMSPConfig) (*MSPConfig, error) {
	mspID := config.Name

	mspConfig := &MSPConfig{
		id:     mspID,
		signer: nil, // no signer when creating from FabricMSPConfig
		users:  nil,
		ouConfig: &OUConfig{
			NodeOUs:         config.FabricNodeOus,
			UnitIdentifiers: config.OrganizationalUnitIdentifiers,
		},
	}

	for _, c := range config.Admins {
		// no key
		adminIdentity, err := FromBytesWithoutSigning(mspID, c)
		if err != nil {
			return nil, fmt.Errorf(`admin cert: %w`, err)
		}

		mspConfig.admins = append(mspConfig.admins, adminIdentity)
	}

	for _, c := range config.RootCerts {
		rootCert, err := Certificate(c)
		if err != nil {
			return nil, fmt.Errorf(`root cert: %w`, err)
		}

		mspConfig.caCerts = append(mspConfig.caCerts, rootCert)
	}

	for _, c := range config.IntermediateCerts {
		intermediateCert, err := Certificate(c)
		if err != nil {
			return nil, fmt.Errorf(`intermediate cert: %w`, err)
		}

		mspConfig.intermediateCerts = append(mspConfig.intermediateCerts, intermediateCert)
	}

	for _, c := range config.TlsRootCerts {
		tlsRootCert, err := Certificate(c)
		if err != nil {
			return nil, fmt.Errorf(`TLS root cert: %w`, err)
		}

		mspConfig.tlsCACerts = append(mspConfig.tlsCACerts, tlsRootCert)
	}

	for _, c := range config.TlsIntermediateCerts {
		tlsIntermediateCert, err := Certificate(c)
		if err != nil {
			return nil, fmt.Errorf(`TLS intermediate cert: %w`, err)
		}

		mspConfig.tlsIntermediateCerts = append(mspConfig.tlsIntermediateCerts, tlsIntermediateCert)
	}

	return mspConfig, nil
}

// MSPFromPath loads msp config from filesystem
func MSPFromPath(mspID, mspPath string, opts ...MspOpt) (*MSPConfig, error) {
	var err error
	mspOpts := &MspOpts{
		mspPath: mspPath,
	}
	for _, opt := range opts {
		opt(mspOpts)
	}

	logger := zap.NewNop()
	if mspOpts.logger != nil {
		logger = mspOpts.logger
	}

	applyDefaultMSPPaths(mspOpts)

	logger.Debug(`load msp`, zap.Reflect(`config`, mspOpts))

	mspConfig := &MSPConfig{
		id: mspID,
	}

	if mspOpts.adminCertsPath != `` {
		mspConfig.admins, err = ListFromPath(mspID, mspOpts.adminCertsPath, mspOpts.keystorePath)
		if err != nil {
			logger.Debug(`load admin identities`, zap.Error(err))
			if !os.IsNotExist(err) {
				return nil, fmt.Errorf(`read admin identity from=%s: %w`, mspOpts.adminCertsPath, err)
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

	if mspOpts.signCertsPath != `` {
		mspConfig.signer, err = FirstFromPath(mspID, mspOpts.signCertsPath, mspOpts.keystorePath)
		if err != nil {
			return nil, fmt.Errorf(`read signer identity from=%s: %w`, mspOpts.signCertsPath, err)
		}
	}

	if mspOpts.loadCertChain {
		mspConfig.caCerts, err = CertificatesFromPath(mspOpts.caCertsPath)
		if err != nil {
			if os.IsNotExist(err) {
				logger.Debug(`cacerts path not found`, zap.String(`path`, mspOpts.caCertsPath))
			} else {
				return nil, fmt.Errorf(`read cacerts from=%s: %w`, mspOpts.caCertsPath, err)
			}
		} else {
			logger.Debug(`CA certs loaded`, zap.Int(`num`, len(mspConfig.caCerts)))
		}

		mspConfig.intermediateCerts, err = CertificatesFromPath(mspOpts.intermediateCertsPath)
		if err != nil {
			if os.IsNotExist(err) {
				logger.Debug(`intermediatecerts path not found`, zap.String(`path`, mspOpts.intermediateCertsPath))
			} else {
				return nil, fmt.Errorf(`read intermediatecerts from=%s: %w`, mspOpts.caCertsPath, err)
			}
		} else {
			logger.Debug(`intermediate certs loaded`, zap.Int(`num`, len(mspConfig.caCerts)))
		}

		mspConfig.tlsCACerts, err = CertificatesFromPath(mspOpts.tlsCaCertsPath)
		if err != nil {
			if os.IsNotExist(err) {
				logger.Debug(`tls cacerts path not found`, zap.String(`path`, mspOpts.tlsCaCertsPath))
			} else {
				return nil, fmt.Errorf(`read tls cacerts from=%s: %w`, mspOpts.tlsCaCertsPath, err)
			}
		} else {
			logger.Debug(`TLS CA certs loaded`, zap.Int(`num`, len(mspConfig.caCerts)))
		}

		mspConfig.tlsIntermediateCerts, err = CertificatesFromPath(mspOpts.tlsIntermediateCertsPath)
		if err != nil {
			if os.IsNotExist(err) {
				logger.Debug(`tls intermediatecerts path not found`, zap.String(`path`, mspOpts.tlsIntermediateCertsPath))
			} else {
				return nil, fmt.Errorf(`read tls intermediatecerts from=%s: %w`, mspOpts.tlsCaCertsPath, err)
			}
		} else {
			logger.Debug(`tls intermediate certs loaded`, zap.Int(`num`, len(mspConfig.caCerts)))
		}

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

// FabricMSPConfig converts info about msp certs and keys to msp.FabricMSPConfig structure
func (m *MSPConfig) FabricMSPConfig() *msp.FabricMSPConfig {
	fabricMSPConfig := &msp.FabricMSPConfig{
		Name:            m.id,
		RevocationList:  nil,
		SigningIdentity: nil,
		CryptoConfig:    nil,
	}

	for _, cert := range m.caCerts {
		fabricMSPConfig.RootCerts = append(fabricMSPConfig.RootCerts, PEMEncode(cert.Raw))
	}

	for _, cert := range m.intermediateCerts {
		fabricMSPConfig.IntermediateCerts = append(fabricMSPConfig.IntermediateCerts, PEMEncode(cert.Raw))
	}

	for _, cert := range m.tlsCACerts {
		fabricMSPConfig.TlsRootCerts = append(fabricMSPConfig.TlsRootCerts, PEMEncode(cert.Raw))
	}

	for _, cert := range m.tlsIntermediateCerts {
		fabricMSPConfig.TlsIntermediateCerts = append(fabricMSPConfig.TlsIntermediateCerts, PEMEncode(cert.Raw))
	}

	for _, adminIdentity := range m.admins {
		fabricMSPConfig.Admins = append(fabricMSPConfig.Admins, adminIdentity.GetPEM())
	}

	if m.ouConfig != nil {
		fabricMSPConfig.OrganizationalUnitIdentifiers = m.ouConfig.UnitIdentifiers
		fabricMSPConfig.FabricNodeOus = m.ouConfig.NodeOUs
	}

	return fabricMSPConfig
}

func (m *MSPConfig) Serialize() (*MSPConfigSerialized, error) {
	return SerializeMSP(m)
}

func (mc MSPCerts) Add(path string, cert []byte) {
	mc[path] = cert
}

func SerializedCertName(path string, pos int) string {
	return fmt.Sprintf(`%s/cert_%d.pem`, path, pos)
}

func SerializeMSP(mspConfig *MSPConfig) (*MSPConfigSerialized, error) {
	var err error

	serialized := &MSPConfigSerialized{
		Certs: make(MSPCerts),
	}

	fabricMSPConfig := mspConfig.FabricMSPConfig()

	for pos, cert := range fabricMSPConfig.Admins {
		serialized.Certs.Add(SerializedCertName(MSPAdminCertsPath, pos), cert)
	}

	for pos, cert := range fabricMSPConfig.RootCerts {
		serialized.Certs.Add(SerializedCertName(MSPCaCertsPath, pos), cert)
	}

	for pos, cert := range fabricMSPConfig.IntermediateCerts {
		serialized.Certs.Add(SerializedCertName(MSPIntermediateCertsPath, pos), cert)
	}

	for pos, cert := range fabricMSPConfig.TlsRootCerts {
		serialized.Certs.Add(SerializedCertName(MSPTLSCaCertsPath, pos), cert)
	}

	for pos, cert := range fabricMSPConfig.TlsIntermediateCerts {
		serialized.Certs.Add(SerializedCertName(MSPTLSIntermediateCertsPath, pos), cert)
	}

	if mspConfig.ouConfig != nil {
		serialized.OU, err = mspConfig.ouConfig.Serialize(MSPOuCertsPath)
		if err != nil {
			return nil, fmt.Errorf(`ou: %w`, err)
		}
	}

	return serialized, nil
}
