package acceptance_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	suiteFailed bool
)

func FailHandler(message string, callerSkip ...int) {
	suiteFailed = true
	Fail(message, callerSkip...)
}

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(FailHandler)
	RunSpecs(t, "Acceptance Suite")
}
