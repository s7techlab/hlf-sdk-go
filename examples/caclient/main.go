package main

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"log"
	mathrand "math/rand"
	"os"
	"time"

	apiCa "github.com/s7techlab/hlf-sdk-go/api/ca"
	"github.com/s7techlab/hlf-sdk-go/ca"
	_ "github.com/s7techlab/hlf-sdk-go/crypto/ecdsa"
	"github.com/s7techlab/hlf-sdk-go/identity"
)

func main() {
	mspId := os.Getenv(`MSP_ID`)
	if mspId == `` {
		log.Fatalln(`MSP_ID env must be defined`)
	}

	configPath := os.Getenv(`CONFIG_PATH`)
	if configPath == `` {
		log.Fatalln(`CONFIG_PATH env must be defined`)
	}

	certPath := os.Getenv(`CERT_PATH`)
	if certPath == `` {
		log.Fatalln(`CERT_PATH env must be defined`)
	}

	keyPath := os.Getenv(`KEY_PATH`)
	if keyPath == `` {
		log.Fatalln(`KEY_PATH env must be defined`)
	}

	id, err := identity.NewMSPIdentity(mspId, certPath, keyPath)
	if err != nil {
		log.Fatalln(`failed to load identity:`, err)
	}

	core, err := ca.NewCore(mspId, id, ca.WithYamlConfig(configPath))
	if err != nil {
		log.Fatalln(`failed to load CA core:`, err)
	}

	log.Println(core.CertificateList(context.Background(), apiCa.WithEnrollId(`admin`)))
	log.Println(core.AffiliationList(context.Background()))
	//log.Println(core.AffiliationCreate(context.Background(), `test`))
	log.Fatalln(``)

	name := `yarrrr` + RandomString(2)

	log.Println(core.Register(apiCa.RegistrationRequest{Name: name, Secret: `123321`}))

	log.Println(core.Enroll(name, `123321`, &x509.CertificateRequest{
		Subject: struct {
			Country, Organization, OrganizationalUnit []string
			Locality, Province                        []string
			StreetAddress, PostalCode                 []string
			SerialNumber, CommonName                  string
			Names                                     []pkix.AttributeTypeAndValue
			ExtraNames                                []pkix.AttributeTypeAndValue
		}{Country: []string{`RU`}, Organization: []string{`S7`}, OrganizationalUnit: []string{`ORG`}, Locality: []string{`Moscow`}, Province: []string{`Moscow`}, StreetAddress: []string{`Пушкина 7`}, PostalCode: []string{`100001`}, CommonName: name},
		SignatureAlgorithm: x509.ECDSAWithSHA512},
	))
}

func RandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func RandomHex(s int) string {
	b := RandomBytes(s)
	hexstring := hex.EncodeToString(b)
	return hexstring
}

func RandomString(s int, letters ...string) string { // s number of character
	randomFactor := RandomBytes(1)
	mathrand.Seed(time.Now().UnixNano() * int64(randomFactor[0]))

	var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	if len(letters) > 0 {
		letterRunes = []rune(letters[0])
	}
	b := make([]rune, s)
	for i := range b {
		b[i] = letterRunes[mathrand.Intn(len(letterRunes))]
	}

	return string(b)
}
