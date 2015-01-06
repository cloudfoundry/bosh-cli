package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMicrodeployer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Installation Manifest Suite")
}
