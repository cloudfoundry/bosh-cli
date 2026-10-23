package ssh

import (
	"crypto/ed25519"
	"crypto/rand"
	"net"
	"os"
	"path/filepath"
	"testing"

	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func newPublicKey(t *testing.T) gossh.PublicKey {
	key, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	public, err := gossh.NewPublicKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return public
}
func TestSSHHostIdentity(t *testing.T) {
	good, bad := newPublicKey(t), newPublicKey(t)
	remote := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22}
	filename := filepath.Join(t.TempDir(), "known_hosts")
	if err := os.WriteFile(filename, []byte(knownhosts.Line([]string{"host.test"}, good)+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, options := range []ClientOpts{{HostPublicKey: string(gossh.MarshalAuthorizedKey(good))}, {KnownHostsFile: filename}} {
		client := &ClientImpl{opts: options}
		verify, err := client.hostKeyCallback()
		if err != nil {
			t.Fatal(err)
		}
		if err := verify("host.test:22", remote, good); err != nil {
			t.Fatalf("trusted key rejected: %v", err)
		}
		mismatch := verify("host.test:22", remote, bad)
		if mismatch == nil {
			t.Fatal("unexpected host key accepted")
		}
		if (&ClientConnectRetryStrategy{}).IsRetryable(mismatch) {
			t.Fatal("host verification failures should fail immediately")
		}
	}
	verify, err := (&ClientImpl{opts: ClientOpts{KnownHostsFile: filename}}).hostKeyCallback()
	if err != nil {
		t.Fatal(err)
	}
	if err := verify("unknown.test:22", remote, good); err == nil {
		t.Fatal("unknown host accepted")
	}
	if _, err := (&ClientImpl{opts: ClientOpts{KnownHostsFile: filename + ".missing"}}).hostKeyCallback(); err == nil {
		t.Fatal("missing trust store accepted")
	}
}
