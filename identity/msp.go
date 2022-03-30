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

	"github.com/s7techlab/hlf-sdk-go/api"
)

// MSP - contains all parsed identities from msp folder
// Should be used instead of single `api.Identity` which contains ONLY msp identity

type (
	MSPConfig struct {
		// identity from 'signcerts'
		signer api.Identity
		// identities from 'admincerts'
		admins []api.Identity
		// identities from 'users'
		users []api.Identity

		mspConfig *mspproto.FabricMSPConfig
	}

	MSPFiles map[string][]byte

	MSP interface {
		GetMSPIdentifier() string
		MSPConfig() *mspproto.FabricMSPConfig

		Signer() api.Identity
		Admins() []api.Identity
		Users() []api.Identity
		AdminOrSigner() api.Identity
	}

	MspOpts struct {
		mspPath string

		adminCertsPath string
		signCertsPath  string
		keystorePath   string

		userPaths []string

		validateCertChain bool
		logger            *zap.Logger
	}

	MspOpt func(opts *MspOpts)
)

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
func MSPFromConfig(fabricMspConfig *mspproto.FabricMSPConfig) (*MSPConfig, error) {
	mspConfig := &MSPConfig{
		admins:    nil,
		signer:    nil, // no signer when creating from FabricMSPConfig
		users:     nil,
		mspConfig: fabricMspConfig,
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

	mspConfig := &MSPConfig{}

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

	if mspConfig.mspConfig, err = FabricMSPConfigFromPath(mspID, mspOpts.mspPath); err != nil {
		return nil, err
	}

	if mspOpts.validateCertChain {
		// todo: validate
	}

	return mspConfig, nil
}

func (m *MSPConfig) GetMSPIdentifier() string {
	return m.mspConfig.GetName()
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

// AdminOrSigner - returns admin identity if exists, in another case return msp.
// installation, fetching  cc list should happen from admin identity
// if there is admin identity, use it. in another case - try with msp identity
func (m *MSPConfig) AdminOrSigner() api.Identity {
	if len(m.admins) != 0 {
		return m.admins[0]
	}

	return m.signer
}

func (m *MSPConfig) MSPConfig() *mspproto.FabricMSPConfig {
	return m.mspConfig
}

func (m *MSPConfig) Serialize() (MSPFiles, error) {
	return SerializeMSP(m.mspConfig)
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
