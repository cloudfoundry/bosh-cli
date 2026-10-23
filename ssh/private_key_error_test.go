package ssh_test

import (
	"strings"
	"testing"

	"github.com/cloudfoundry/bosh-cli/v7/ssh"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

func TestPrivateKeyParseErrorOmitsInput(t *testing.T) {
	client := ssh.NewClientFactory(boshlog.NewLogger(boshlog.LevelNone)).New(ssh.ClientOpts{PrivateKey: "private-key-fragment"})
	err := client.Start()
	if err == nil || strings.Contains(err.Error(), "private-key-fragment") {
		t.Fatalf("unsafe parse error: %v", err)
	}
}
