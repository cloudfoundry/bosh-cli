package deviceutil_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDeviceutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Deviceutil Suite")
}
