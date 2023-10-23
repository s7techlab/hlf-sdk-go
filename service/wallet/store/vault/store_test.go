// // +build unit
package vault_test

//
//import (
//	. "github.com/onsi/ginkgo"
//	. "github.com/onsi/gomega"
//
//	"errors"
//	"net/http"
//	"os"
//	"time"
//
//	"github.com/hashicorp/vault/api"
//	"github.com/ory/dockertest/v3"
//	"github.com/ory/dockertest/v3/docker"
//
//	"b2bchain.tech/pkg/hlf/wallet"
//)
//
//const (
//	vaultToken = "root"
//)
//
//var (
//	_, gitlabCIEnv = os.LookupEnv(`CI`)
//	vaultAddress   string
//	pool           *dockertest.Pool
//	vault          *dockertest.Resource
//)
//
//var _ = BeforeSuite(func(done Done) {
//
//	if gitlabCIEnv {
//		vaultAddress = os.Getenv("VAULT_ADDR")
//		close(done)
//		return
//	}
//
//	By("creating dockertest pool")
//
//	var err error
//
//	pool, err = dockertest.NewPool("")
//	Expect(err).ShouldNot(HaveOccurred())
//
//	By("running vault in dockertest pool")
//
//	opts := dockertest.RunOptions{
//		Repository: "vault",
//		Tag:        "latest",
//		Env: []string{
//			"VAULT_DEV_ROOT_TOKEN_ID=" + vaultToken,
//		},
//		ExposedPorts: []string{"8200"},
//		PortBindings: map[docker.Port][]docker.PortBinding{
//			"8200": {
//				{HostIP: "127.0.0.1", HostPort: "8200"},
//			},
//		},
//	}
//
//	vault, err = pool.RunWithOptions(&opts)
//	Expect(err).ShouldNot(HaveOccurred())
//
//	vaultAddress = "http://127.0.0.1:8200"
//
//	// wait vault docker to start
//	time.Sleep(time.Second)
//
//	close(done)
//
//}, 100)
//
//var _ = AfterSuite(func() {
//	if !gitlabCIEnv && vault != nil {
//		By("purging vault from dockertest pool")
//		err := pool.Purge(vault)
//		Expect(err).ShouldNot(HaveOccurred())
//	}
//})
//
//var _ = Describe("StoreHashicorpVault", func() {
//	Context("Get", func() {
//		It("should return right IdentityInWallet", func() {
//
//			By("creating vault API client")
//
//			c, err := api.NewClient(&api.Config{
//				Address: vaultAddress,
//			})
//			Expect(err).ShouldNot(HaveOccurred())
//
//			c.SetToken("root")
//
//			By("preparing POST request")
//
//			req := c.NewRequest(http.MethodPost, "/v1/secret/data/foo/bar")
//
//			expV := &wallet.IdentityInWallet{
//				Label:        "bar",
//				MspId:        "test_mspid",
//				Cert:         []byte("test_cert"),
//				Key:          []byte("test_key"),
//				WithPassword: false,
//			}
//
//			err = req.SetJSONBody(map[string]interface{}{"data": expV})
//			Expect(err).ShouldNot(HaveOccurred())
//
//			By("making POST request")
//
//			res, err := c.RawRequest(req)
//			Expect(err).ShouldNot(HaveOccurred())
//			if !(res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNoContent) {
//				Fail("unexpected response status code")
//			}
//
//			By("creating HashicorpVaultStore")
//
//			s, err := wallet.NewVaultStore(vaultAddress, "/foo", vaultToken)
//			Expect(err).ShouldNot(HaveOccurred())
//
//			By("getting IdentityInWallet")
//
//			gotV, err := s.Get("bar")
//			Expect(err).ShouldNot(HaveOccurred())
//
//			By("comparing got IdentityInWallet with expected")
//
//			Expect(gotV).Should(Equal(expV))
//		})
//	})
//
//	Context("Set", func() {
//		It("should successfully set IdentityInWallet", func() {
//
//			By("creating HashicorpVaultStore")
//
//			s, err := wallet.NewVaultStore(vaultAddress, "/foo2", vaultToken)
//			Expect(err).ShouldNot(HaveOccurred())
//
//			By("setting expected IdentityInWallet")
//
//			expV := &wallet.IdentityInWallet{
//				Label:        "bar",
//				MspId:        "test_mspid_bar",
//				Cert:         []byte("test_cert_bar"),
//				Key:          []byte("test_key_bar"),
//				WithPassword: false,
//			}
//
//			err = s.Set(expV)
//			Expect(err).ShouldNot(HaveOccurred())
//
//			By("getting IdentityInWallet")
//
//			gotV, err := s.Get("bar")
//			Expect(err).ShouldNot(HaveOccurred())
//
//			By("comparing got IdentityInWallet with expected")
//
//			Expect(gotV).Should(Equal(expV))
//		})
//	})
//	Context("List", func() {
//		It("should successfully list labels", func() {
//
//			By("creating HashicorpVaultStore")
//
//			s, err := wallet.NewVaultStore(vaultAddress, "/foo3", vaultToken)
//			Expect(err).ShouldNot(HaveOccurred())
//
//			By("setting IdentityInWallet with label `bar`")
//
//			err = s.Set(&wallet.IdentityInWallet{
//				Label:        "bar",
//				MspId:        "test_mspid_bar",
//				Cert:         []byte("test_cert_bar"),
//				Key:          []byte("test_key_bar"),
//				WithPassword: false,
//			})
//			Expect(err).ShouldNot(HaveOccurred())
//
//			By("setting IdentityInWallet with label `bar2`")
//
//			err = s.Set(&wallet.IdentityInWallet{
//				Label:        "bar2",
//				MspId:        "test_mspid_bar2",
//				Cert:         []byte("test_cert_bar2"),
//				Key:          []byte("test_key_bar2"),
//				WithPassword: false,
//			})
//			Expect(err).ShouldNot(HaveOccurred())
//
//			By("listing labels")
//
//			gotLabels, err := s.List()
//			Expect(err).ShouldNot(HaveOccurred())
//
//			By("comparing got labels with expected")
//
//			expLabels := []string{
//				"bar",
//				"bar2",
//			}
//
//			Expect(gotLabels).Should(Equal(expLabels))
//		})
//	})
//
//	Context("Delete", func() {
//		It("should successfully delete", func() {
//
//			By("creating HashicorpVaultStore")
//
//			s, err := wallet.NewVaultStore(vaultAddress, "/foo4", vaultToken)
//			Expect(err).ShouldNot(HaveOccurred())
//
//			By("setting IdentityInWallet with label `bar`")
//
//			err = s.Set(&wallet.IdentityInWallet{
//				Label:        "bar",
//				MspId:        "test_mspid_bar",
//				Cert:         []byte("test_cert_bar"),
//				Key:          []byte("test_key_bar"),
//				WithPassword: false,
//			})
//			Expect(err).ShouldNot(HaveOccurred())
//
//			By("Deleting")
//
//			err = s.Delete("bar")
//			Expect(err).ShouldNot(HaveOccurred())
//
//			By("comparing got labels with expected")
//
//			_, err = s.Get("bar")
//			Expect(errors.Is(err, wallet.ErrIdentityNotFound)).To(BeTrue())
//		})
//	})
//})
