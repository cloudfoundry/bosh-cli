package property_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestProperty(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Common Property Suite")
}
