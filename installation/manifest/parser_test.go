package manifest_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	. "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
)

var _ = Describe("Parser", func() {
	var (
		comboManifestPath string
		fakeFs            *fakesys.FakeFileSystem
		parser            Parser
	)

	BeforeEach(func() {
		comboManifestPath = "fake-deployment-path"
		fakeFs = fakesys.NewFakeFileSystem()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		parser = NewParser(fakeFs, logger)

		contents := `
---
name: fake-deployment-name
cloud_provider:
  ssh_tunnel:
    host: 54.34.56.8
    port: 22
    user: fake-ssh-user
    private_key: /tmp/fake-ssh-key.pem
  agent_env_service: registry
  mbus: http://fake-mbus-user:fake-mbus-password@0.0.0.0:6868
  registry:
    username: fake-registry-username
    password: fake-registry-password
    host: fake-registry-host
    port: 123
  properties:
    fake-property-name:
      nested-property: fake-property-value
`
		fakeFs.WriteFileString(comboManifestPath, contents)
	})

	Context("when combo manifest path does not exist", func() {
		BeforeEach(func() {
			err := fakeFs.RemoveAll(comboManifestPath)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error", func() {
			_, err := parser.Parse(comboManifestPath)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when parser fails to read the combo manifest file", func() {
		BeforeEach(func() {
			fakeFs.ReadFileError = errors.New("fake-read-file-error")
		})

		It("returns an error", func() {
			_, err := parser.Parse(comboManifestPath)
			Expect(err).To(HaveOccurred())
		})
	})

	It("parses installation from combo manifest", func() {
		installationManifest, err := parser.Parse(comboManifestPath)
		Expect(err).ToNot(HaveOccurred())

		Expect(installationManifest).To(Equal(Manifest{
			Name: "fake-deployment-name",
			Registry: Registry{
				Username: "fake-registry-username",
				Password: "fake-registry-password",
				Host:     "fake-registry-host",
				Port:     123,
			},
			AgentEnvService: "registry",
			RawProperties: map[interface{}]interface{}{
				"fake-property-name": map[interface{}]interface{}{
					"nested-property": "fake-property-value",
				},
			},
			SSHTunnel: SSHTunnel{
				Host:       "54.34.56.8",
				Port:       22,
				User:       "fake-ssh-user",
				PrivateKey: "/tmp/fake-ssh-key.pem",
			},
			Mbus: "http://fake-mbus-user:fake-mbus-password@0.0.0.0:6868",
		}))
	})
})
