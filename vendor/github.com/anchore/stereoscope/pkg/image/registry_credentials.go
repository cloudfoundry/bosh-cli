package image

import (
	"github.com/anchore/stereoscope/internal/log"
	"github.com/google/go-containerregistry/pkg/authn"
)

// RegistryCredentials contains any information necessary to authenticate against an OCI-distribution-compliant
// registry (either with basic auth or bearer token). Note: only valid for the OCI registry provider.
type RegistryCredentials struct {
	Authority string
	Username  string
	Password  string
	Token     string
}

// authenticator returns an authn.Authenticator for the given credentials.
// Authentication methods are attempted in the following order until a viable method is found: (1) basic auth,
// (2) bearer token. If no viable authentication method is found, authenticator returns nil.
func (c RegistryCredentials) authenticator() authn.Authenticator {
	if c.Username != "" && c.Password != "" {
		log.Debugf("using basic auth for registry %q", c.Authority)
		return &authn.Basic{
			Username: c.Username,
			Password: c.Password,
		}
	}

	if c.Token != "" {
		log.Debugf("using token for registry %q", c.Authority)
		return &authn.Bearer{
			Token: c.Token,
		}
	}

	return nil
}

// canBeUsedWithRegistry returns a bool indicating if these credentials should be used when accessing the given registry.
func (c RegistryCredentials) canBeUsedWithRegistry(registry string) bool {
	if !c.hasAuthoritySpecified() {
		return true
	}

	return registry == c.Authority
}

// hasAuthoritySpecified returns a bool indicating if there is a specified "authority" value,
// meaning that the user has requested these credentials to be used for retrieving only the images whose registry
// matches this "authority" value.
func (c RegistryCredentials) hasAuthoritySpecified() bool {
	return c.Authority != ""
}
