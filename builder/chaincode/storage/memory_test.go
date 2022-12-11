package storage

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"

	"github.com/s7techlab/hlf-sdk-go/builder/chaincode"
)

var (
	m   *memory
	ctx = context.Background()
)

var _ = Describe("Memory storage", func() {

	It("allow to init", func() {
		m = NewMemory()
		Expect(m).NotTo(BeNil())
	})

	It("allow to put item", func() {
		err := m.Put(ctx, chaincode.Package{
			PackageID: chaincode.PackageID{
				Name:          "cc",
				Version:       "v1",
				FabricVersion: chaincode.FabricV2,
			},
			Repository:    "path://to.repo",
			ChaincodePath: "cc/path/",
			BinaryPath:    "bin",
			Data:          []byte(`aaaaaaaaasonmebytes`),
		})

		Expect(err).NotTo(HaveOccurred())
	})

	It("allow to get item", func() {
		pi, err := m.Get(ctx, chaincode.PackageID{
			Name:          "cc",
			Version:       "v1",
			FabricVersion: chaincode.FabricV2,
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(pi.Name).To(Equal("cc"))
		Expect(pi.Version).To(Equal("v1"))
		Expect(pi.FabricVersion).To(Equal(chaincode.FabricV2))
	})

	It("should returns error for unknown package", func() {
		_, err := m.Get(ctx, chaincode.PackageID{
			Name:          "cc_other",
			Version:       "v1",
			FabricVersion: chaincode.FabricV2,
		})

		Expect(err).To(HaveOccurred())
	})

	It("allow to fetch item multiple times", func() {
		for i := 0; i < 3; i++ {
			data, err := m.Fetch(ctx, chaincode.PackageID{
				Name:          "cc",
				Version:       "v1",
				FabricVersion: chaincode.FabricV2,
			})

			Expect(err).NotTo(HaveOccurred())

			bytes, err := ioutil.ReadAll(data)

			Expect(err).NotTo(HaveOccurred())
			Expect(bytes).To(Equal([]byte(`aaaaaaaaasonmebytes`)))
		}
	})
})
