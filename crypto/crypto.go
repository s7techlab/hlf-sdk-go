package crypto

import (
	"crypto/rand"
	"sync"

	"github.com/pkg/errors"

	"github.com/atomyze-ru/hlf-sdk-go/api"
	"github.com/atomyze-ru/hlf-sdk-go/api/config"
)

var (
	suiteRegistry   = make(map[string]api.CryptoSuite)
	suiteMx         sync.Mutex
	errUnknownSuite = errors.New(`unknown crypto suite (forgotten import ?)`)
)

// Register must be called in init function in suite package (ex. see ecdsa package)
func Register(name string, cs api.CryptoSuite) {
	suiteMx.Lock()
	defer suiteMx.Unlock()
	suiteRegistry[name] = cs
}

func GetSuite(name string, opts config.CryptoSuiteOpts) (api.CryptoSuite, error) {
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
