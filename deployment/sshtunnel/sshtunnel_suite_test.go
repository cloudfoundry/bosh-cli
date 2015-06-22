package sshtunnel

import (
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega"
	"testing"
)

func TestSshtunnel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sshtunnel Suite")
}
