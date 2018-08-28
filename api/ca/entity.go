package ca

type (
	Identity struct {
		Id             string              `json:"id"`
		Type           string              `json:"type"`
		MaxEnrollments int                 `json:"max_enrollments"`
		Name           string              `json:"name"`
		Attrs          []IdentityAttribute `json:"attrs"`
	}

	IdentityAttribute struct {
		Name  string `json:"name"`
		Value string `json:"value"`
		Ecert bool   `json:"ecert"`
	}

	RevokedCert struct {
		Serial string
		AKI    string
	}
)
