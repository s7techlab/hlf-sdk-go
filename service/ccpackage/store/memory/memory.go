package memory

import (
	"bytes"
	"context"
	"io"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"

	"github.com/s7techlab/hlf-sdk-go/service/ccpackage"
	"github.com/s7techlab/hlf-sdk-go/service/ccpackage/store"
)

const (
	// 	memoryPackageTTL package in memory time to live
	memoryPackageTTL   = 24 * time.Hour
	memoryTickerPeriod = 1 * time.Hour
)

type Storage struct {
	packages map[string]*ccpackage.Package
	mx       sync.RWMutex
	wg       sync.WaitGroup
	stop     chan struct{}
}

func New() *Storage {
	m := &Storage{
		packages: map[string]*ccpackage.Package{},
		stop:     make(chan struct{}),
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

			for key, pi := range m.packages {
				if time.Since(pi.CreatedAt.AsTime()) > memoryPackageTTL {
					delete(m.packages, key)
				}
			}

			m.mx.Unlock()
		}
	}()

	return m
}

// Put saves chaincode package into packages.
func (m *Storage) Put(ctx context.Context, req *ccpackage.PutPackageRequest) error {
	m.mx.Lock()
	defer m.mx.Unlock()

	m.packages[store.ObjectKey(req.Id)] = &ccpackage.Package{
		Id:        req.Id,
		Size:      int64(len(req.Data)),
		CreatedAt: ptypes.TimestampNow(),
		Data:      req.Data,
	}
	return nil
}

// Get gets chaincode package info from packages.
func (m *Storage) Get(_ context.Context, id *ccpackage.PackageID) (*ccpackage.Package, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	value, ok := m.packages[store.ObjectKey(id)]
	if !ok {
		return nil, store.ErrPackageNotFound
	}
	return value, nil
}

// List gets stored chaincode packages' infos.
func (m *Storage) List(ctx context.Context) ([]*ccpackage.Package, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	var pgs []*ccpackage.Package
	for _, p := range m.packages {
		pgs = append(pgs, &ccpackage.Package{
			Id:        p.Id,
			Size:      p.Size,
			CreatedAt: p.CreatedAt,
		})
	}

	return pgs, nil
}

// Fetch fetches chaincode package.
func (m *Storage) Fetch(ctx context.Context, id *ccpackage.PackageID) (io.ReadCloser, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	pkg, err := m.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewBuffer(pkg.Data)), nil
}

func (m *Storage) Close() error {
	close(m.stop)
	m.wg.Wait()
	return nil
}
