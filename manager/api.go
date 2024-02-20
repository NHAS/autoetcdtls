package manager

const (
	AuthHeader = "X-AUTH"
	// Path to get the public key of the cluster CA
	getCACert = "/public/fetch/cert"

	// Path to get the private key of the cluster CA. This path is served over TLS  and requires a password in X-AUTH
	getCAPrivateKey = "/private/fetch/key"

	// Leave room for additionals
	getAdditionals = "/private/additionals"
)
