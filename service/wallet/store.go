package wallet

type (
	Store interface {
		Get(label string) (*IdentityInWallet, error)
		Set(identity *IdentityInWallet) error
		List() (labels []string, err error)
		Delete(label string) error
	}
)
