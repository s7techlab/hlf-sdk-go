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
		ECert bool   `json:"ecert"`
	}

	RevokedCert struct {
		Serial string
		AKI    string
	}

	Affiliation struct {
		Name         string        `json:"name"`
		Affiliations []Affiliation `json:"affiliations,omitempty"`
		Identities   []Identity    `json:"identities,omitempty"`
	}
)
