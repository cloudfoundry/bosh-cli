package instance_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestInstance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Instance Suite")
}
