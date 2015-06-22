package http_test

import (
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega"
	"testing"
)

func TestAgentclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTP AgentClient Suite")
}
