package memory

import (
	"github.com/s7techlab/hlf-sdk-go/proto/wallet"
	wallet2 "github.com/s7techlab/hlf-sdk-go/service/wallet"
)

type (
	MemoryStore struct {
		identities map[string]*wallet.IdentityInWallet
	}
)

func New() *MemoryStore {
	return &MemoryStore{
		identities: make(map[string]*wallet.IdentityInWallet),
	}
}

func (ms *MemoryStore) Get(label string) (*wallet.IdentityInWallet, error) {
	id, exists := ms.identities[label]
	if !exists {
		return nil, wallet2.ErrIdentityNotFound
	}

	return id, nil
}

func (ms *MemoryStore) Set(identity *wallet.IdentityInWallet) error {
	ms.identities[identity.Label] = identity
	return nil
}

func (ms *MemoryStore) List() ([]string, error) {
	var labels []string

	for l := range ms.identities {
		labels = append(labels, l)
	}

	return labels, nil
}

func (ms *MemoryStore) Delete(label string) error {
	delete(ms.identities, label)
	return nil
}
