package vm_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVM(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "VM Suite")
}
