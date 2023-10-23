package memory

import (
	wallet2 "github.com/s7techlab/hlf-sdk-go/service/wallet"
)

type (
	MemoryStore struct {
		identities map[string]*wallet2.IdentityInWallet
	}
)

func New() *MemoryStore {
	return &MemoryStore{
		identities: make(map[string]*wallet2.IdentityInWallet),
	}
}

func (ms *MemoryStore) Get(label string) (*wallet2.IdentityInWallet, error) {
	id, exists := ms.identities[label]
	if !exists {
		return nil, wallet2.ErrIdentityNotFound
	}

	return id, nil
}

func (ms *MemoryStore) Set(identity *wallet2.IdentityInWallet) error {
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
