package keystringifier_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestKeystringifier(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Keystringifier Suite")
}
