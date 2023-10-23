package crypto

import (
	"crypto/rand"
	"errors"
	"sync"

	"github.com/s7techlab/hlf-sdk-go/api/config"
	"github.com/s7techlab/hlf-sdk-go/crypto/ecdsa"
)

var (
	suiteRegistry   = make(map[string]CryptoSuite)
	suiteMx         sync.Mutex
	errUnknownSuite = errors.New(`unknown crypto suite (forgotten import ?)`)
)

// Register must be called in init function in suite package (ex. see ecdsa package)
func Register(name string, cs CryptoSuite) {
	suiteMx.Lock()
	defer suiteMx.Unlock()
	suiteRegistry[name] = cs
}

func GetSuite(name string, opts config.CryptoSuiteOpts) (CryptoSuite, error) {
	suiteMx.Lock()
	defer suiteMx.Unlock()
	if suite, ok := suiteRegistry[name]; ok {
		return suite.Initialize(opts)
	}
	return nil, errUnknownSuite
}

// RandomBytes returns slice of random bytes of presented size
func RandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func DefaultCryptoSuite() CryptoSuite {
	suite, _ := GetSuite(ecdsa.Module, ecdsa.DefaultOpts)
	return suite
}
