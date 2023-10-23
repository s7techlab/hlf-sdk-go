package crypto

import (
	"crypto/rand"
)

// RandomBytes returns slice of random bytes of presented size
func RandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
