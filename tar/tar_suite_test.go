package tar_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTar(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tar Suite")
}
