package identity

import (
	"fmt"
	"os"
	"path"

	"github.com/golang/protobuf/proto"
	mspproto "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric/msp"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// MSP - contains all parsed identities from msp folder
// Should be used instead of single `api.Identity` which contains ONLY msp identity

type (
	MSP struct {
		// identity from 'signcerts'
		signer *SigningIdentity
		// identities from 'admincerts'
		admins []*SigningIdentity
		// identities from 'users'
		users []*SigningIdentity

		config *mspproto.FabricMSPConfig
	}

	MSPFiles map[string][]byte

	//MSP interface {
	//	GetMSPIdentifier() string
	//	MSPConfig() *mspproto.FabricMSPConfig
	//
	//	Signer() *SigningIdentity
	//	Admins() []*SigningIdentity
	//	Users() []*SigningIdentity
	//	AdminOrSigner() *SigningIdentity
	//}

	MSPOpts struct {
		mspPath string

		// signCert and signKey take precedence over signCertPath and signKeyPath
		signCert []byte
		signKey  []byte

		signCertPath string
		signKeyPath  string

		signCertsPath  string
		keystorePath   string
		adminCertsPath string
		adminMSPPath   string

		userPaths []string

		skipConfig bool
		logger     *zap.Logger
	}

	MSPOpt func(opts *MSPOpts)
)

func applyDefaultMSPPaths(mspOpts *MSPOpts) {
	if mspOpts.adminCertsPath == `` {
		mspOpts.adminCertsPath = AdminCertsPath(mspOpts.mspPath)
	}

	if mspOpts.signCertsPath == `` {
		mspOpts.signCertsPath = SignCertsPath(mspOpts.mspPath)
	}

	if mspOpts.keystorePath == `` {
		mspOpts.keystorePath = KeystorePath(mspOpts.mspPath)
	}
}

func FabricMSPConfigFromPath(mspID, mspDir string) (*mspproto.FabricMSPConfig, error) {
	serializedConfig, err := msp.GetLocalMspConfig(mspDir, nil, mspID)
	if err != nil {
		return nil, fmt.Errorf(`get local msp config from path=%s: %w`, mspDir, err)
	}

	// double marshal/unmarshal
	mspConfig := &mspproto.FabricMSPConfig{}
	err = proto.Unmarshal(serializedConfig.Config, mspConfig)

	return mspConfig, err
}

// MSPFromConfig created  msp config from msp.FabricMSPConfig
func MSPFromConfig(fabricMspConfig *mspproto.FabricMSPConfig) (*MSP, error) {
	mspInstance := &MSP{
		admins: nil,
		signer: nil, // no signer when creating from FabricMSPConfig
		users:  nil,
		config: fabricMspConfig,
	}
	return mspInstance, nil
}

// MSPFromPath loads msp config from filesystem
func MSPFromPath(mspID, mspPath string, opts ...MSPOpt) (*MSP, error) {
	var err error
	mspOpts := &MSPOpts{
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

	mspInstance := &MSP{}

	if len(mspOpts.signCert) != 0 && len(mspOpts.signKey) != 0 {
		mspInstance.signer, err = NewSigningFromBytes(mspID, mspOpts.signCert, mspOpts.signKey)
		if err != nil {
			return nil, err
		}
	} else if mspOpts.signCertPath != "" && mspOpts.signKeyPath != "" {
		mspInstance.signer, err = NewSigningFromFile(mspID, mspOpts.signCertPath, mspOpts.signKeyPath)
		if err != nil {
			return nil, err
		}
	}

	// admin in separate msp path
	if mspOpts.adminMSPPath != `` {
		logger.Debug(`load admin identities from separate msp path`,
			zap.String(`admin msp path`, mspOpts.adminMSPPath),
			zap.String(`keystore path`, KeystorePath(mspOpts.adminMSPPath)))

		mspInstance.admins, err = ListSigningFromPath(mspID, SignCertsPath(mspOpts.adminMSPPath), KeystorePath(mspOpts.adminMSPPath))

		if err != nil {
			return nil, fmt.Errorf(`read admin identity from=%s: %w`, mspOpts.adminMSPPath, err)
		}
	} else if mspOpts.adminCertsPath != `` {
		mspInstance.admins, err = ListSigningFromPath(mspID, mspOpts.adminCertsPath, mspOpts.keystorePath)
	}
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf(`read admin identity from=%s: %w`, mspOpts.adminCertsPath, err)
		}
	}

	logger.Debug(`admin identities loaded`, zap.Int(`num`, len(mspInstance.admins)))

	if len(mspOpts.userPaths) > 0 {
		for _, userPath := range mspOpts.userPaths {
			users, err := ListSigningFromPath(mspID, userPath, mspOpts.keystorePath)
			// usePaths set explicit, so if dir is not exists - error occurred
			if err != nil {
				return nil, fmt.Errorf(`read users identity from=%s: %w`, userPath, err)
			}

			mspInstance.users = append(mspInstance.users, users...)
		}

		logger.Debug(`user identities loaded`, zap.Int(`num`, len(mspInstance.users)))
	}

	if mspOpts.signCertsPath != `` && mspInstance.signer == nil {
		mspInstance.signer, err = FirstSigningFromPath(mspID, mspOpts.signCertsPath, mspOpts.keystorePath)
		if err != nil {
			return nil, fmt.Errorf(`read signer identity from=%s: %w`, mspOpts.signCertsPath, err)
		}
	}

	if !mspOpts.skipConfig {
		if mspInstance.config, err = FabricMSPConfigFromPath(mspID, mspOpts.mspPath); err != nil {
			return nil, err
		}
	}

	// todo: validate
	//if mspOpts.validateCertChain {
	//}

	return mspInstance, nil
}

func (m *MSP) Identifier() string {
	return m.config.GetName()
}

func (m *MSP) Signer() *SigningIdentity {
	return m.signer
}

func (m *MSP) Admins() []*SigningIdentity {
	return m.admins
}

func (m *MSP) Users() []*SigningIdentity {
	return m.users
}

// AdminOrSigner - returns admin identity if exists, in another case return msp.
// installation, fetching  cc list should happen from admin identity
// if there is admin identity, use it. in another case - try with msp identity
func (m *MSP) AdminOrSigner() *SigningIdentity {
	if len(m.admins) != 0 {
		return m.admins[0]
	}

	return m.signer
}

func (m *MSP) Config() *mspproto.FabricMSPConfig {
	return m.config
}

func (m *MSP) Serialize() (MSPFiles, error) {
	return SerializeMSP(m.config)
}

func (mc MSPFiles) Add(path string, file []byte) {
	mc[path] = file
}

func (mc MSPFiles) Merge(files MSPFiles) {
	for filePath, file := range files {
		mc[filePath] = file
	}
}

func (mc MSPFiles) MergeToPath(mergePath string, files MSPFiles) {
	for filePath, file := range files {
		mc[path.Join(mergePath, filePath)] = file
	}
}

func SerializedCertName(path string, pos int) string {
	return fmt.Sprintf(`%s/cert_%d.pem`, path, pos)
}

func SerializeMSP(fabricMSPConfig *mspproto.FabricMSPConfig) (MSPFiles, error) {
	files := make(MSPFiles)

	for pos, cert := range fabricMSPConfig.Admins {
		files.Add(SerializedCertName(MSPAdminCertsPath, pos), cert)
	}

	for pos, cert := range fabricMSPConfig.RootCerts {
		files.Add(SerializedCertName(MSPCaCertsPath, pos), cert)
	}

	for pos, cert := range fabricMSPConfig.IntermediateCerts {
		files.Add(SerializedCertName(MSPIntermediateCertsPath, pos), cert)
	}

	for pos, cert := range fabricMSPConfig.TlsRootCerts {
		files.Add(SerializedCertName(MSPTLSCaCertsPath, pos), cert)
	}

	for pos, cert := range fabricMSPConfig.TlsIntermediateCerts {
		files.Add(SerializedCertName(MSPTLSIntermediateCertsPath, pos), cert)
	}

	var (
		clientOUFile  = path.Join(MSPOuCertsPath, `client.pem`)
		peerOUFile    = path.Join(MSPOuCertsPath, `peer.pem`)
		adminOUFile   = path.Join(MSPOuCertsPath, `admin.pem`)
		ordererOUFile = path.Join(MSPOuCertsPath, `orderer.pem`)
	)

	if nodeOUs := fabricMSPConfig.FabricNodeOus; nodeOUs != nil && nodeOUs.Enable {

		mspConfig := &msp.Configuration{}

		mspConfig.NodeOUs = &msp.NodeOUs{
			Enable: nodeOUs.Enable,
		}

		if nodeOUs.ClientOuIdentifier != nil && nodeOUs.ClientOuIdentifier.OrganizationalUnitIdentifier != `` {
			mspConfig.NodeOUs.ClientOUIdentifier = &msp.OrganizationalUnitIdentifiersConfiguration{
				OrganizationalUnitIdentifier: nodeOUs.ClientOuIdentifier.OrganizationalUnitIdentifier,
			}

			if len(nodeOUs.ClientOuIdentifier.Certificate) != 0 {
				files.Add(clientOUFile, nodeOUs.ClientOuIdentifier.Certificate)
				mspConfig.NodeOUs.ClientOUIdentifier.Certificate = clientOUFile
			}
		}

		if nodeOUs.PeerOuIdentifier != nil && nodeOUs.PeerOuIdentifier.OrganizationalUnitIdentifier != `` {
			mspConfig.NodeOUs.PeerOUIdentifier = &msp.OrganizationalUnitIdentifiersConfiguration{
				OrganizationalUnitIdentifier: nodeOUs.PeerOuIdentifier.OrganizationalUnitIdentifier,
			}

			if len(nodeOUs.PeerOuIdentifier.Certificate) != 0 {
				files.Add(peerOUFile, nodeOUs.PeerOuIdentifier.Certificate)
				mspConfig.NodeOUs.PeerOUIdentifier.Certificate = peerOUFile
			}
		}

		if nodeOUs.AdminOuIdentifier != nil && nodeOUs.AdminOuIdentifier.OrganizationalUnitIdentifier != `` {
			mspConfig.NodeOUs.AdminOUIdentifier = &msp.OrganizationalUnitIdentifiersConfiguration{
				OrganizationalUnitIdentifier: nodeOUs.AdminOuIdentifier.OrganizationalUnitIdentifier,
			}

			if len(nodeOUs.AdminOuIdentifier.Certificate) != 0 {
				files.Add(adminOUFile, nodeOUs.AdminOuIdentifier.Certificate)
				mspConfig.NodeOUs.AdminOUIdentifier.Certificate = adminOUFile
			}
		}

		if nodeOUs.OrdererOuIdentifier != nil && nodeOUs.OrdererOuIdentifier.OrganizationalUnitIdentifier != `` {
			mspConfig.NodeOUs.OrdererOUIdentifier = &msp.OrganizationalUnitIdentifiersConfiguration{
				OrganizationalUnitIdentifier: nodeOUs.OrdererOuIdentifier.OrganizationalUnitIdentifier,
			}

			if len(nodeOUs.OrdererOuIdentifier.Certificate) != 0 {
				files.Add(ordererOUFile, nodeOUs.OrdererOuIdentifier.Certificate)
				mspConfig.NodeOUs.OrdererOUIdentifier.Certificate = ordererOUFile
			}
		}

		config, err := yaml.Marshal(mspConfig)
		if err != nil {
			return nil, fmt.Errorf(`marshal config.yaml: %w`, err)
		}

		files.Add(MspConfigFile, config)
	}

	return files, nil
}
