package ca

import (
	"encoding/json"
)

type (
	Response struct {
		Success  bool              `json:"success"`
		Result   json.RawMessage   `json:"result"`
		Errors   []ResponseMessage `json:"errors"`
		Messages []ResponseMessage `json:"messages"`
	}

	ResponseMessage struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	ResponseCAInfo struct {
		CAName  string `json:"CAName"`
		CAChain string `json:"CAChain"`
		Version string `json:"Version"`
	}

	ResponseRegistration struct {
		Secret string `json:"secret"`
	}

	ResponseEnrollment struct {
		Cert       string         `json:"Cert"`
		ServerInfo ResponseCAInfo `json:"ServerInfo"`
	}

	ResponseIdentityList struct {
		Identities []Identity `json:"identities"`
	}

	ResponseCertificateList struct {
		CAName string                       `json:"caname"`
		Certs  []ResponseCertificateListPEM `json:"certs"`
	}

	ResponseCertificateListPEM struct {
		PEM string `json:"PEM"`
	}

	ResponseRevoke struct {
		RevokedCerts []RevokedCert
		CRL          []byte
	}

	ResponseAffiliationList struct {
		Name         string        `json:"name"`
		Affiliations []Affiliation `json:"affiliations"`
		Identities   []Identity    `json:"identities"`
		CAName       string        `json:"caname"`
	}

	ResponseAffiliationCreate struct {
		Name   string `json:"name"`
		CAName string `json:"caname"`
	}

	ResponseAffiliationDelete struct {
		ResponseAffiliationList
	}
)
