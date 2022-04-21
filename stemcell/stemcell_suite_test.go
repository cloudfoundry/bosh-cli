package stemcell_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestStemcell(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stemcell Suite")
}
