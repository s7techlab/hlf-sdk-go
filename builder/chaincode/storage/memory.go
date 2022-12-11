package storage

import (
	"bytes"
	"context"
	"io"
	"sync"
	"time"

	"github.com/s7techlab/hlf-sdk-go/builder/chaincode"
)

const (
	memoryPackageTTL   = 24 * time.Hour
	memoryTickerPeriod = 1 * time.Hour
)

type memory struct {
	storage map[chaincode.PackageInfo]*bytes.Buffer
	mx      sync.RWMutex
	wg      sync.WaitGroup
	stop    chan struct{}
}

func NewMemory() *memory {
	m := &memory{
		storage: map[chaincode.PackageInfo]*bytes.Buffer{},
		stop:    make(chan struct{}),
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		t := time.NewTicker(memoryTickerPeriod)
		defer t.Stop()

		for {
			select {
			case <-m.stop:
				return
			case <-t.C:
			}

			m.mx.Lock()

			for pi := range m.storage {
				if time.Since(pi.CreatedAt) > memoryPackageTTL {
					delete(m.storage, pi)
				}
			}

			m.mx.Unlock()
		}
	}()

	return m
}

// Put saves chaincode package into storage.
func (m *memory) Put(ctx context.Context, pkg chaincode.Package) error {

	m.mx.Lock()
	defer m.mx.Unlock()

	m.storage[chaincode.PackageInfo{
		PackageID: pkg.PackageID,
		Size:      len(pkg.Data),
		CreatedAt: time.Now(),
	}] = bytes.NewBuffer(pkg.Data)

	return nil
}

// Get gets chaincode package info from storage.
func (m *memory) Get(ctx context.Context, id chaincode.PackageID) (
	chaincode.PackageInfo, error) {

	m.mx.RLock()
	defer m.mx.RUnlock()

	for pi := range m.storage {
		if pi.PackageID == id {
			return pi, nil
		}
	}

	return chaincode.PackageInfo{}, ErrPackageNotFound
}

// List gets stored chaincode packages' infos.
func (m *memory) List(ctx context.Context) ([]chaincode.PackageInfo, error) {

	m.mx.RLock()
	defer m.mx.RUnlock()

	var pis []chaincode.PackageInfo
	for pi := range m.storage {
		pis = append(pis, pi)
	}

	return pis, nil
}

// Fetch fetches chaincode package.
func (m *memory) Fetch(ctx context.Context, id chaincode.PackageID) (io.ReadCloser, error) {

	m.mx.RLock()
	defer m.mx.RUnlock()

	for pi, p := range m.storage {
		if pi.PackageID == id {
			return io.NopCloser(bytes.NewBuffer(p.Bytes())), nil
		}
	}

	return nil, ErrPackageNotFound
}

func (m *memory) Close() error {
	close(m.stop)
	m.wg.Wait()
	return nil
}
