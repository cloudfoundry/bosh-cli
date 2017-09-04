package windows_test

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/bosh-agent/integration/windows/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestWindows(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Windows Suite")
}

var vagrantProvider = os.Getenv("VAGRANT_PROVIDER")

var _ = BeforeSuite(func() {
	if _, ok := os.LookupEnv("NATS_PRIVATE_IP"); !ok {
		Fail("Environment variable NATS_PRIVATE_IP not set", 1)
	}
	_, err := utils.StartVagrant(vagrantProvider)
	if err != nil {
		Fail(fmt.Sprintln("Could not build the bosh-agent project.\nError is:", err))
	}
})
