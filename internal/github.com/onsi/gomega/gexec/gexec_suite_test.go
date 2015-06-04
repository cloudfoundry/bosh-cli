package gexec_test

import (
	"github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega/gexec"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var fireflyPath string

func TestGexec(t *testing.T) {
	BeforeSuite(func() {
		var err error
		fireflyPath, err = gexec.Build("./_fixture/firefly")
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "Gexec Suite")
}
