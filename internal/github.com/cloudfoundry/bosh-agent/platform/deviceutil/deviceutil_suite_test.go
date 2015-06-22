package deviceutil_test

import (
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega"

	"testing"
)

func TestDeviceutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Deviceutil Suite")
}
