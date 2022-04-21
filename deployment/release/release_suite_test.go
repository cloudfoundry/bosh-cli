package release_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Deployment Release Suite")
}
