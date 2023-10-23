package vault

import (
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/hashicorp/vault/api"

	wallet2 "github.com/s7techlab/hlf-sdk-go/service/wallet"
)

type (
	Store struct {
		client *api.Client
		prefix string
	}

	getData struct {
		Data struct {
			IdentityInWallet *wallet2.IdentityInWallet `json:"data"`
		} `json:"data"`
	}

	setData struct {
		IdentityInWallet *wallet2.IdentityInWallet `json:"data"`
	}

	listData struct {
		Data struct {
			Labels []string `json:"keys"`
		} `json:"data"`
	}
)

const (
	defaultPrefix     = "/v1/secret/data"
	defaultListPrefix = "/v1/secret/metadata"
	methodList        = "LIST"
)

func NewVault(connection string) (*Store, error) {
	parsedURL, err := url.Parse(connection)
	if err != nil {
		return nil, err
	}

	c, err := api.NewClient(&api.Config{
		Address: parsedURL.Scheme + `://` + parsedURL.Host,
	})
	if err != nil {
		return nil, err
	}
	c.SetToken(parsedURL.User.Username())

	return &Store{client: c, prefix: parsedURL.Path}, nil
}

func (s *Store) Get(label string) (*wallet2.IdentityInWallet, error) {
	req := s.client.NewRequest(http.MethodGet, path.Join(defaultPrefix, s.prefix, label))

	res, err := s.client.RawRequest(req)
	if err != nil {

		if res != nil && res.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf(`%s: %w`, err, wallet2.ErrIdentityNotFound)
		}

		return nil, fmt.Errorf("make request: %w", err)
	}

	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status code=%d", res.StatusCode)
	}

	data := new(getData)

	err = res.DecodeJSON(&data)
	if err != nil {
		return nil, fmt.Errorf("decode response data: %w", err)
	}

	return data.Data.IdentityInWallet, nil
}

//
//func (s *VaultStore) Set(identity *IdentityInWallet) error {
//	req := s.client.NewRequest(http.MethodPost, path.Join(defaultPrefix, s.prefix, identity.Label))
//
//	err := req.SetJSONBody(setData{
//		IdentityInWallet: identity,
//	})
//	if err != nil {
//		return fmt.Errorf("set request JSON body: %w", err)
//	}
//
//	res, err := s.client.RawRequest(req)
//	if err != nil {
//		return fmt.Errorf("make request: %w", err)
//	}
//
//	if !(res.StatusCode == http.StatusNoContent || res.StatusCode == http.StatusOK) {
//		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
//	}
//
//	defer func() { _ = res.Body.Close() }()
//
//	return nil
//}
//
//func (s *VaultStore) List() ([]string, error) {
//	req := s.client.NewRequest(methodList, path.Join(defaultListPrefix, s.prefix))
//
//	res, err := s.client.RawRequest(req)
//	if err != nil {
//		return nil, fmt.Errorf("make request: %w", err)
//	}
//
//	defer func() { _ = res.Body.Close() }()
//
//	if res.StatusCode != http.StatusOK {
//		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
//	}
//
//	var data listData
//
//	err = json.NewDecoder(res.Body).Decode(&data)
//	if err != nil {
//		return nil, fmt.Errorf("decode labels: %w", err)
//	}
//
//	return data.Data.Labels, nil
//}
//
//func (s *VaultStore) Delete(label string) error {
//	req := s.client.NewRequest(http.MethodDelete, path.Join(defaultPrefix, s.prefix, label))
//	res, err := s.client.RawRequest(req)
//	if err != nil {
//		return fmt.Errorf("delete request: %w", err)
//	}
//
//	defer func() { _ = res.Body.Close() }()
//
//	if res.StatusCode != http.StatusNoContent {
//		body, _ := ioutil.ReadAll(res.Body)
//		return fmt.Errorf("delete request result: %d, %s", res.StatusCode, string(body))
//	}
//
//	return nil
//}
