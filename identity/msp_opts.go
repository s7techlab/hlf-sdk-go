package identity

func WithSkipConfig() MSPOpt {
	return func(mspOpts *MSPOpts) {
		mspOpts.skipConfig = true
	}
}

func WithAdminMSPPath(adminMSPPath string) MSPOpt {
	return func(mspOpts *MSPOpts) {
		mspOpts.adminMSPPath = adminMSPPath
	}
}

func WithSignCertPath(signCertPath string) MSPOpt {
	return func(mspOpts *MSPOpts) {
		mspOpts.signCertPath = signCertPath
	}
}

func WithSignKeyPath(signKeyPath string) MSPOpt {
	return func(mspOpts *MSPOpts) {
		mspOpts.signKeyPath = signKeyPath
	}
}

func WithSignCert(signCert []byte) MSPOpt {
	return func(mspOpts *MSPOpts) {
		mspOpts.signCert = signCert
	}
}

func WithSignKey(signKey []byte) MSPOpt {
	return func(mspOpts *MSPOpts) {
		mspOpts.signKey = signKey
	}
}
