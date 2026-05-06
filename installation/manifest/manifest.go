package manifest

import (
	"crypto/x509"

	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
)

type Manifest struct {
	Name string
	// Deprecated: use Templates instead
	Template   ReleaseJobRef
	Templates  []ReleaseJobRef
	Properties biproperty.Map
	Mbus       string
	Cert       Certificate
}

type Certificate struct {
	CA string
}

// CACertPool parses the PEM-encoded CA certificate and returns an *x509.CertPool.
// Returns (nil, nil) when CA is empty, meaning the system root CAs will be used.
func (c Certificate) CACertPool() (*x509.CertPool, error) {
	if len(c.CA) == 0 {
		return nil, nil
	}
	return boshcrypto.CertPoolFromPEM([]byte(c.CA))
}

type ReleaseJobRef struct {
	Name    string
	Release string
}

type SSHTunnel struct {
	User       string
	Host       string
	Port       int
	Password   string
	PrivateKey string `yaml:"private_key"`
}
