package manifest_test

import (
	"errors"
	"os/user"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/installation/manifest"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	biproperty "github.com/cloudfoundry/bosh-init/common/property"
)

var _ = Describe("Parser", func() {
	var (
		comboManifestPath string
		fakeFs            *fakesys.FakeFileSystem
		parser            Parser
		logger            boshlog.Logger
	)

	BeforeEach(func() {
		fakeFs = fakesys.NewFakeFileSystem()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		parser = NewParser(fakeFs, logger)
		comboManifestPath = "fake-deployment-manifest"
	})

	Context("when combo manifest path does not exist", func() {
		It("returns an error", func() {
			_, err := parser.Parse(comboManifestPath)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when parser fails to read the combo manifest file", func() {
		JustBeforeEach(func() {
			fakeFs.WriteFileString(comboManifestPath, "---\n")
			fakeFs.ReadFileError = errors.New("fake-read-file-error")
		})

		It("returns an error", func() {
			_, err := parser.Parse(comboManifestPath)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the combo manifest file refrences ~", func() {
		Context("when the ~ begins the path", func() {
			It("expands ~/ into the current user's home directory", func() {
				manifest := `
---
name: fake-deployment-name
cloud_provider:
  template:
    name: fake-cpi-job-name
    release: fake-cpi-release-name
  ssh_tunnel:
    host: 54.34.56.8
    port: 22
    user: fake-ssh-user
    private_key: ~/fake-ssh-key.pem
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
				manifestPath := "4d89da25-454a-4c96-9f70-7bc9b90fee53"
				fakeFs.WriteFileString(manifestPath, manifest)

				currentUser, err := user.Current()
				Expect(err).ToNot(HaveOccurred())

				installationManifest, err := parser.Parse(manifestPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(installationManifest.SSHTunnel.PrivateKey).To(ContainSubstring(currentUser.HomeDir))
			})
		})

		Context("when ~ occurs anywhere else in the path", func() {
			It("assumes the user desired a literal ~", func() {
				manifest := `
---
name: fake-deployment-name
cloud_provider:
  template:
    name: fake-cpi-job-name
    release: fake-cpi-release-name
  ssh_tunnel:
    host: 54.34.56.8
    port: 22
    user: fake-ssh-user
    private_key: /some-root/~/fake-ssh-key.pem
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
				currentUser, err := user.Current()
				Expect(err).ToNot(HaveOccurred())

				manifestPath := "1a56d58a-0d56-4743-bbc5-5a9056af8bed"
				fakeFs.WriteFileString(manifestPath, manifest)

				installationManifest, err := parser.Parse(manifestPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(installationManifest.SSHTunnel.PrivateKey).ToNot(ContainSubstring(currentUser.HomeDir))
			})
		})
	})

	Context("with a valid manifest", func() {
		JustBeforeEach(func() {
			contents := `
---
name: fake-deployment-name
cloud_provider:
  template:
    name: fake-cpi-job-name
    release: fake-cpi-release-name
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

		It("parses installation from combo manifest", func() {
			installationManifest, err := parser.Parse(comboManifestPath)
			Expect(err).ToNot(HaveOccurred())

			Expect(installationManifest).To(Equal(Manifest{
				Name: "fake-deployment-name",
				Template: ReleaseJobRef{
					Name:    "fake-cpi-job-name",
					Release: "fake-cpi-release-name",
				},
				Registry: Registry{
					Username: "fake-registry-username",
					Password: "fake-registry-password",
					Host:     "fake-registry-host",
					Port:     123,
				},
				AgentEnvService: "registry",
				Properties: biproperty.Map{
					"fake-property-name": biproperty.Map{
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
})
