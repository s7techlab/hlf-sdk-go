package ca

type (
	// RegistrationRequest holds all data needed for new registration of new user in Certificate Authority
	RegistrationRequest struct {
		// Name is unique name that identifies identity
		Name string `json:"name"`
		// Type defines type of this identity (user,client, auditor etc...)
		Type string `json:"type"`
		// Secret is password that will be used for enrollment. If not provided random password will be generated
		Secret string `json:"secret,omitempty"`
		// MaxEnrollments define maximum number of times that identity can enroll. If not provided or is 0 there is no limit
		MaxEnrollments int `json:"max_enrollments,omitempty"`
		// Affiliation associates identity with particular organisation.
		// for example org1.department1 makes this identity part of organisation `org1` and department `department1`
		// Hierarchical structure can be created using .(dot). For example org1.dep1 will create dep1 as part of org1
		Affiliation string `json:"affiliation"`
		// Attrs are attributes associated with this identity
		Attrs []RegisterAttribute `json:"attrs"`
		// CAName is the name of the CA that should be used. FabricCa support more than one CA server on same endpoint and
		// this names are used to distinguish between them. If empty default CA instance will be used.
		CAName string `json:"caname,omitempty"`
	}

	// RegisterAttribute holds user attribute used for registration
	// for example user may have attr `accountType` with value `premium`
	// this attributes can be accessed in chainCode and build business logic on top of them
	RegisterAttribute struct {
		// Name is the name of the attribute.
		Name string `json:"name"`
		// Value is the value of the attribute. Can be empty string
		Value string `json:"value"`
		// ECert define how this attribute will be included in ECert. If this value is true this attribute will be
		// added to ECert automatically on Enrollment if no attributes are requested on Enrollment request.
		ECert bool `json:"ecert,omitempty"`
	}

	// EnrollmentRequest holds data needed for getting ECert (enrollment) from CA server
	EnrollmentRequest struct {
		// EnrollmentId is the unique entity identifies
		EnrollmentId string
		// Secret is the password for this identity
		Secret string
		// Profile define which CA profile to be used for signing. When this profile is empty default profile is used.
		// This is the common situation when issuing and ECert.
		// If request is fo generating TLS certificates then profile must be `tls`
		// If operation is related to parent CA server then profile must be `ca`
		// In FabricCA custom profiles can be created. In this situation use custom profile name.
		Profile string `json:"profile,omitempty"`
		// Label is used for hardware secure modules.
		Label string `json:"label,omitempty"`
		// CAName is the name of the CA that should be used. FabricCa support more than one CA server on same endpoint and
		// this names are used to distinguish between them. If empty default CA instance will be used.
		CAName string `json:"caname,omitempty"`
		// Host is the list of valid host names for this certificate. If empty default hosts will be used
		Hosts []string `json:"hosts"`
		// Attrs are the attributes that must be included in ECert. This is subset of the attributes used in registration.
		Attrs []EnrollAttribute `json:"attr_reqs,omitempty"`
	}

	// ReEnrollmentRequest holds data needed for getting new ECert from CA server
	ReEnrollmentRequest struct {
		// Profile define which CA profile to be used for signing. When this profile is empty default profile is used.
		// This is the common situation when issuing and ECert.
		// If request is fo generating TLS certificates then profile must be `tls`
		// If operation is related to parent CA server then profile must be `ca`
		// In FabricCA custom profiles can be created. In this situation use custom profile name.
		Profile string `json:"profile,omitempty"`
		// Label is used for hardware secure modules.
		Label string `json:"label,omitempty"`
		// CAName is the name of the CA that should be used. FabricCa support more than one CA server on same endpoint and
		// this names are used to distinguish between them. If empty default CA instance will be used.
		CAName string `json:"caname,omitempty"`
		// Host is the list of valid host names for this certificate. If empty default hosts will be used
		Hosts []string `json:"hosts"`
		// Attrs are the attributes that must be included in ECert. This is subset of the attributes used in registration.
		Attrs []EnrollAttribute `json:"attr_reqs,omitempty"`
	}

	// EnrollAttribute describe attribute that must be included in enrollment request
	EnrollAttribute struct {
		// Name is the name of the attribute
		Name string `json:"name"`
		// Optional define behaviour when required attribute is not available to user. If `true` then request will continue,
		// but attribute will not be included in ECert. If `false` and attribute is missing, request will fail.
		// If false and attribute is available, request will continue and attribute will be added in ECert
		Optional bool `json:"optional,omitempty"`
	}
)
