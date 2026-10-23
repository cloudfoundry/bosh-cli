package ssh

import (
	"net"
	"os"
	"path/filepath"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type hostKeyVerificationError struct{ cause error }

func (e hostKeyVerificationError) Error() string { return "verifying SSH host key: " + e.cause.Error() }
func (e hostKeyVerificationError) Unwrap() error { return e.cause }

func (s *ClientImpl) hostKeyCallback() (gossh.HostKeyCallback, error) {
	var callback gossh.HostKeyCallback
	if s.opts.HostPublicKey != "" {
		key, _, _, _, err := gossh.ParseAuthorizedKey([]byte(s.opts.HostPublicKey))
		if err != nil {
			return nil, bosherr.WrapError(err, "Parsing SSH host public key")
		}
		callback = gossh.FixedHostKey(key)
	} else {
		filename := s.opts.KnownHostsFile
		if filename == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, bosherr.WrapError(err, "Finding SSH known hosts")
			}
			filename = filepath.Join(homeDir, ".ssh", "known_hosts")
		}
		var err error
		callback, err = knownhosts.New(filename)
		if err != nil {
			return nil, bosherr.WrapError(err, "Loading SSH known hosts")
		}
	}
	return func(host string, remote net.Addr, key gossh.PublicKey) error {
		if err := callback(host, remote, key); err != nil {
			return hostKeyVerificationError{err}
		}
		return nil
	}, nil
}
