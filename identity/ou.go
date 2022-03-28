package identity

import (
	"fmt"
	"os"
	"path"

	protomsp "github.com/hyperledger/fabric-protos-go/msp"
	"github.com/hyperledger/fabric/msp"
	"gopkg.in/yaml.v2"
)

type OUConfig struct {
	NodeOUs         *protomsp.FabricNodeOUs
	UnitIdentifiers []*protomsp.FabricOUIdentifier
}

func ReadOUIDConfig(dir string, ouIDConfig *msp.OrganizationalUnitIdentifiersConfiguration) (*protomsp.FabricOUIdentifier, error) {
	var err error

	ouID := &protomsp.FabricOUIdentifier{
		OrganizationalUnitIdentifier: ouIDConfig.OrganizationalUnitIdentifier,
	}

	if ouID.Certificate, err = readOuCertificate(dir, ouIDConfig); err != nil {
		return nil, err
	}

	return ouID, err
}

// ReadNodeOUConfig Load configuration file
// if the configuration file is there then load it
// otherwise skip it
func ReadNodeOUConfig(dir string) (*OUConfig, error) {
	configRaw, err := readConfig(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}

	var (
		configuration msp.Configuration
		ouConfig      OUConfig
	)

	err = yaml.Unmarshal(configRaw, &configuration)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling config.yaml: %w", err)
	}

	// Prepare OrganizationalUnitIdentifiers
	for _, ouIDConfig := range configuration.OrganizationalUnitIdentifiers {
		ouID, ouErr := ReadOUIDConfig(dir, ouIDConfig)
		if ouErr != nil {
			return nil, ouErr
		}
		ouConfig.UnitIdentifiers = append(ouConfig.UnitIdentifiers, ouID)
	}

	// Prepare NodeOUs
	if configuration.NodeOUs != nil && configuration.NodeOUs.Enable {
		ouConfig.NodeOUs = &protomsp.FabricNodeOUs{
			Enable: true,
		}

		if configuration.NodeOUs.ClientOUIdentifier != nil && len(configuration.NodeOUs.ClientOUIdentifier.OrganizationalUnitIdentifier) != 0 {
			if ouConfig.NodeOUs.ClientOuIdentifier, err = ReadOUIDConfig(dir, configuration.NodeOUs.ClientOUIdentifier); err != nil {
				return nil, err
			}
		}

		if configuration.NodeOUs.PeerOUIdentifier != nil && len(configuration.NodeOUs.PeerOUIdentifier.OrganizationalUnitIdentifier) != 0 {
			if ouConfig.NodeOUs.PeerOuIdentifier, err = ReadOUIDConfig(dir, configuration.NodeOUs.PeerOUIdentifier); err != nil {
				return nil, err
			}
		}
		if configuration.NodeOUs.AdminOUIdentifier != nil && len(configuration.NodeOUs.AdminOUIdentifier.OrganizationalUnitIdentifier) != 0 {
			if ouConfig.NodeOUs.AdminOuIdentifier, err = ReadOUIDConfig(dir, configuration.NodeOUs.AdminOUIdentifier); err != nil {
				return nil, err
			}
		}
		if configuration.NodeOUs.OrdererOUIdentifier != nil && len(configuration.NodeOUs.OrdererOUIdentifier.OrganizationalUnitIdentifier) != 0 {
			if ouConfig.NodeOUs.OrdererOuIdentifier, err = ReadOUIDConfig(dir, configuration.NodeOUs.OrdererOUIdentifier); err != nil {
				return nil, err
			}
		}
	}

	return &ouConfig, nil
}

func SerializeOU(certPath string, ouConfig *OUConfig) (MSPFiles, error) {

	files := make(MSPFiles)
	const (
		clientOUFile  = `client.pem`
		peerOUFile    = `peer.pem`
		adminOUFile   = `admin.pem`
		ordererOUFile = `orderer.pem`
	)

	mspConfig := &msp.Configuration{
		NodeOUs: &msp.NodeOUs{
			Enable: ouConfig.NodeOUs.Enable,
		},
	}

	if nodeOUs := ouConfig.NodeOUs; nodeOUs != nil {

		if nodeOUs.ClientOuIdentifier != nil && nodeOUs.ClientOuIdentifier.OrganizationalUnitIdentifier != `` {
			mspConfig.NodeOUs.ClientOUIdentifier = &msp.OrganizationalUnitIdentifiersConfiguration{
				OrganizationalUnitIdentifier: nodeOUs.ClientOuIdentifier.OrganizationalUnitIdentifier,
			}

			if len(nodeOUs.ClientOuIdentifier.Certificate) != 0 {
				files.Add(clientOUFile, nodeOUs.ClientOuIdentifier.Certificate)
				mspConfig.NodeOUs.ClientOUIdentifier.Certificate = path.Join(certPath, clientOUFile)
			}
		}

		if nodeOUs.PeerOuIdentifier != nil && nodeOUs.PeerOuIdentifier.OrganizationalUnitIdentifier != `` {
			mspConfig.NodeOUs.PeerOUIdentifier = &msp.OrganizationalUnitIdentifiersConfiguration{
				OrganizationalUnitIdentifier: nodeOUs.PeerOuIdentifier.OrganizationalUnitIdentifier,
			}

			if len(nodeOUs.PeerOuIdentifier.Certificate) != 0 {
				files.Add(peerOUFile, nodeOUs.PeerOuIdentifier.Certificate)
				mspConfig.NodeOUs.PeerOUIdentifier.Certificate = path.Join(certPath, peerOUFile)
			}
		}

		if nodeOUs.AdminOuIdentifier != nil && nodeOUs.AdminOuIdentifier.OrganizationalUnitIdentifier != `` {
			mspConfig.NodeOUs.AdminOUIdentifier = &msp.OrganizationalUnitIdentifiersConfiguration{
				OrganizationalUnitIdentifier: nodeOUs.AdminOuIdentifier.OrganizationalUnitIdentifier,
			}

			if len(nodeOUs.AdminOuIdentifier.Certificate) != 0 {
				files.Add(adminOUFile, nodeOUs.AdminOuIdentifier.Certificate)
				mspConfig.NodeOUs.AdminOUIdentifier.Certificate = path.Join(certPath, adminOUFile)
			}
		}

		if nodeOUs.OrdererOuIdentifier != nil && nodeOUs.OrdererOuIdentifier.OrganizationalUnitIdentifier != `` {
			mspConfig.NodeOUs.OrdererOUIdentifier = &msp.OrganizationalUnitIdentifiersConfiguration{
				OrganizationalUnitIdentifier: nodeOUs.OrdererOuIdentifier.OrganizationalUnitIdentifier,
			}

			if len(nodeOUs.OrdererOuIdentifier.Certificate) != 0 {
				files.Add(ordererOUFile, nodeOUs.OrdererOuIdentifier.Certificate)
				mspConfig.NodeOUs.OrdererOUIdentifier.Certificate = path.Join(certPath, ordererOUFile)
			}
		}
	}

	config, err := yaml.Marshal(mspConfig)
	if err != nil {
		return nil, fmt.Errorf(`marshal config.yaml: %w`, err)
	}

	files.Add(MspConfigFile, config)

	return files, nil
}

func (ou *OUConfig) Serialize(certPath string) (MSPFiles, error) {
	return SerializeOU(certPath, ou)
}
