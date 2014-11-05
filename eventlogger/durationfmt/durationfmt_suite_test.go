package durationfmt_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDurationFmt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DurationFmt Suite")
}
