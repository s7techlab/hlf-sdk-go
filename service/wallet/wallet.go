package wallet

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"errors"
	"fmt"
	"regexp"

	empty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/s7techlab/hlf-sdk-go/proto/wallet"
	"github.com/s7techlab/hlf-sdk-go/service"
)

var (
	ErrEmptyLabel         = errors.New(`empty label`)
	ErrInvalidCharInLabel = errors.New(`invalid char in label`)

	DisallowedCharsInLabel, _ = regexp.Compile("[^A-Za-z0-9_-]+")
)

type (
	Wallet struct {
		store Store
	}
)

func New(store Store) *Wallet {
	return &Wallet{
		store: store,
	}
}

func (w *Wallet) ServiceDef() *service.Def {
	return service.NewDef(
		`wallet`,
		wallet.Swagger,
		&wallet.ServiceDesc,
		w,
		wallet.RegisterWalletServiceHandlerFromEndpoint)
}

func ValidateLabel(label string) error {
	if label == `` {
		return ErrEmptyLabel
	}
	if DisallowedCharsInLabel.Match([]byte(label)) {
		return ErrInvalidCharInLabel
	}
	return nil
}

func (w *Wallet) IdentityGet(_ context.Context, lbl *wallet.IdentityLabel) (*wallet.IdentityInWallet, error) {
	if err := ValidateLabel(lbl.Label); err != nil {
		return nil, err
	}

	id, err := w.store.Get(lbl.Label)
	if err != nil {
		return nil, fmt.Errorf(`label = %s: %w`, lbl.Label, err)
	}

	return id, nil
}

func (w *Wallet) IdentityGetText(ctx context.Context, lbl *wallet.IdentityLabel) (
	*wallet.IdentityInWalletText, error) {
	id, err := w.IdentityGet(ctx, lbl)
	if err != nil {
		return nil, err
	}

	//var content string
	//cert, err := x509util.CertificateFromPEM(id.Cert)
	//
	//if err != nil {
	//	content = err.Error()
	//} else {
	//	content = x509util.CertificateToString(cert)
	//}

	return &wallet.IdentityInWalletText{
		Label: id.Label,
		MspId: id.MspId,
		Role:  id.Role,
		Cert:  string(id.Cert),
		//CertContent:  content,
		Key:          string(id.Key),
		WithPassword: id.WithPassword,
	}, nil
}

func (w *Wallet) IdentitySet(ctx context.Context, identity *wallet.Identity) (*wallet.IdentityInWallet, error) {
	if err := ValidateLabel(identity.Label); err != nil {
		return nil, err
	}

	identityInWallet := &wallet.IdentityInWallet{
		Label:        identity.Label,
		MspId:        identity.MspId,
		Role:         identity.Role,
		Cert:         identity.Cert,
		Key:          identity.Key,
		WithPassword: false,
	}

	if err := w.store.Set(identityInWallet); err != nil {
		return nil, err
	}

	return identityInWallet, nil
}

func (w *Wallet) IdentitySetWithPassword(_ context.Context, identity *wallet.IdentityWithPassword) (
	*wallet.IdentityInWallet, error) {
	encryptedPemBlock, err := x509.EncryptPEMBlock(rand.Reader, `EC PRIVATE KEY`, identity.Key,
		[]byte(identity.Password), x509.PEMCipherAES256)
	if err != nil {
		return nil, fmt.Errorf("encrypt pem block: %w", err)
	}

	encryptedKey := pem.EncodeToMemory(encryptedPemBlock)

	identityInWallet := &wallet.IdentityInWallet{
		Label:        identity.Label,
		MspId:        identity.MspId,
		Role:         identity.Role,
		Cert:         identity.Cert,
		Key:          encryptedKey,
		WithPassword: true,
	}

	if err := w.store.Set(identityInWallet); err != nil {
		return nil, fmt.Errorf("set in store: %w", err)
	}

	return identityInWallet, nil
}

func (w *Wallet) IdentityGetWithPassword(_ context.Context, identity *wallet.IdentityPassword) (*wallet.IdentityInWallet, error) {
	if err := ValidateLabel(identity.Label); err != nil {
		return nil, err
	}

	identityInWallet, err := w.store.Get(identity.Label)
	if err != nil {
		return nil, fmt.Errorf(`label = %s: %w`, identity.Label, err)
	}

	if !identityInWallet.WithPassword {
		return nil, fmt.Errorf("identity is without password")
	}

	pemBlock, _ := pem.Decode(identityInWallet.Key)
	if pemBlock == nil {
		return nil, fmt.Errorf("pem decode key: %w", err)
	}

	decryptedKey, err := x509.DecryptPEMBlock(pemBlock, []byte(identity.Password))
	if err != nil {
		return nil, fmt.Errorf("decrypt pem block: %w", err)
	}

	identityInWallet.Key = decryptedKey
	identityInWallet.WithPassword = false

	return identityInWallet, nil
}

func (w *Wallet) IdentityList(context.Context, *empty.Empty) (*wallet.IdentityLabels, error) {
	labels, err := w.store.List()
	if err != nil {
		return nil, err
	}

	return &wallet.IdentityLabels{
		Labels: labels,
	}, nil
}

func (w *Wallet) IdentityDelete(ctx context.Context, label *wallet.IdentityLabel) (*wallet.IdentityInWallet, error) {
	id, err := w.IdentityGet(ctx, label)
	if err != nil {
		return nil, err
	}

	if err = w.store.Delete(label.Label); err != nil {
		return nil, err
	}

	return id, nil
}
