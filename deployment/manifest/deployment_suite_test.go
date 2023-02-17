package manifest_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDeployment(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Deployment Suite")
}
