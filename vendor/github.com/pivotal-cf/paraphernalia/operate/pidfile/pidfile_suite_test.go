package pidfile_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPidfile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pidfile Suite")
}
