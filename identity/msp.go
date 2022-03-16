package identity

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/s7techlab/hlf-sdk-go/api"
)

var (
	ErrSignCertsDirIsEmpty = errors.New(`"signcerts"" folder is empty`)
)

// MSP - contains all parsed identities from msp folder
// Should be used instead of single `api.Identity` which contains ONLY msp identity
type MSP struct {
	// identity from 'signcerts'
	MSP api.Identity
	// identities from 'admincerts'
	Admins []api.Identity
	// identities from 'users'
	Users []api.Identity
}

// CollectionFromMSPPath  parse all certificates(msp,admins,users) from MSP folder.
// Came to replace legacy `util.LoadKeyPairFromMSP` method
func CollectionFromMSPPath(mspID string, mspPath string) (*MSP, error) {

	mspIdentities := &MSP{
		Admins: make([]api.Identity, 0),
		MSP:    nil,
		Users:  make([]api.Identity, 0),
	}

	// admin certs
	adminDir := filepath.Join(mspPath, MSPAdmincertsPath)
	_, err := os.Stat(adminDir)
	if !os.IsNotExist(err) {
		adminCerts, err := ReadAllFilesFromDir(adminDir)
		if err != nil {
			return nil, err
		}

		for _, adminCertBytes := range adminCerts {
			cert, key, err := LoadKeypairByCert(mspPath, adminCertBytes)
			if err != nil {
				return nil, err
			}

			idnt, _ := New(mspID, cert, key)
			mspIdentities.Admins = append(mspIdentities.Admins, idnt)
		}
	}

	// user certs
	userDir := filepath.Join(mspPath, MSPUserscertsPath)
	_, err = os.Stat(userDir)
	if !os.IsNotExist(err) {
		userCerts, err := ReadAllFilesFromDir(userDir)
		if err != nil {
			return nil, err
		}

		for _, userCertBytes := range userCerts {
			cert, key, err := LoadKeypairByCert(mspPath, userCertBytes)
			if err != nil {
				return nil, err
			}

			idnt, _ := New(mspID, cert, key)
			mspIdentities.Users = append(mspIdentities.Users, idnt)
		}
	}

	// signcert
	signCertsDir := filepath.Join(mspPath, MSPSigncertsPath)
	signCerts, err := ReadAllFilesFromDir(signCertsDir)
	if err != nil {
		return nil, err
	}
	if len(signCerts) == 0 {
		return nil, ErrSignCertsDirIsEmpty
	}

	cert, key, err := LoadKeypairByCert(mspPath, signCerts[0])
	if err != nil {
		return nil, err
	}

	idnt, _ := New(mspID, cert, key)
	mspIdentities.MSP = idnt

	return mspIdentities, nil
}

// PreferAdminIfExists - returns admin identity if exists, in another case return msp.
// installation, fetching  cc list should happen from admin identity
// if there is admin identity, use it. in another case - try with msp identity
func (m *MSP) PreferAdminIfExists() api.Identity {
	var identity api.Identity
	if len(m.Admins) != 0 {
		identity = m.Admins[0]
	} else {
		identity = m.MSP
	}
	return identity
}
