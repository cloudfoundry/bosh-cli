package util_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProperty(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Common Util Suite")
}
