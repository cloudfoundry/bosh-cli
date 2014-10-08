package microdeployer_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMicrodeployer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Microdeployer Suite")
}
