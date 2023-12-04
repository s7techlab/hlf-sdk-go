package memory_test

import (
	"context"
	"io"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/s7techlab/hlf-sdk-go/service/ccpackage"
	"github.com/s7techlab/hlf-sdk-go/service/ccpackage/store"
	"github.com/s7techlab/hlf-sdk-go/service/ccpackage/store/memory"
)

func TestMemoryStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Memory storage test suite")
}

var (
	m *memory.Storage

	packageID1 = &ccpackage.PackageID{
		Name:          "cc",
		Version:       "v1",
		FabricVersion: ccpackage.FabricVersion_FABRIC_V2,
	}

	packageData1 = []byte(`aaaaaaaaasonmebytes`)
)

var _ = Describe("Memory package storage", func() {

	ctx := context.Background()

	It("allow to init", func() {
		m = memory.New()
		Expect(m).NotTo(BeNil())
	})

	It("allow to put item", func() {
		err := m.Put(ctx, &ccpackage.PutPackageRequest{
			Id:   packageID1,
			Data: packageData1,
		})

		Expect(err).NotTo(HaveOccurred())
	})

	It("allow to get item", func() {
		p, err := m.Get(ctx, packageID1)

		Expect(err).NotTo(HaveOccurred())
		Expect(p.Id.String()).To(Equal(packageID1.String()))
		Expect(int(p.Size)).To(Equal(len(packageData1)))
	})

	It("should returns error for unknown package", func() {
		_, err := m.Get(ctx, &ccpackage.PackageID{
			Name:          "xxx",
			Version:       "xxx",
			FabricVersion: 0,
		})
		Expect(err).To(MatchError(store.ErrPackageNotFound))
	})

	It("allow to fetch item multiple times", func() {
		for i := 0; i < 3; i++ {
			data, err := m.Fetch(ctx, packageID1)
			Expect(err).NotTo(HaveOccurred())

			bytes, err := io.ReadAll(data)

			Expect(err).NotTo(HaveOccurred())
			Expect(bytes).To(Equal(packageData1))
		}
	})
})
