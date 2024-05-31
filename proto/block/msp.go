package block

// GetAllCertificates - returns all certificates from MSP
func (x *MSP) GetAllCertificates() ([]*Certificate, error) {
	var certs []*Certificate

	for i := range x.Config.RootCerts {
		cert, err := NewCertificate(x.Config.RootCerts[i], CertType_CERT_TYPE_CA, x.Config.Name, x.Name)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}

	for i := range x.Config.IntermediateCerts {
		cert, err := NewCertificate(x.Config.IntermediateCerts[i], CertType_CERT_TYPE_INTERMEDIATE, x.Config.Name, x.Name)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}

	for i := range x.Config.Admins {
		cert, err := NewCertificate(x.Config.Admins[i], CertType_CERT_TYPE_ADMIN, x.Config.Name, x.Name)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}

	return certs, nil
}
