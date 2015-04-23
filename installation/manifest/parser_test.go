package manifest_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/installation/manifest"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	biproperty "github.com/cloudfoundry/bosh-init/common/property"
)

var _ = Describe("Parser", func() {
	var (
		comboManifestPath string
		fakeFs            *fakesys.FakeFileSystem
		fakeUUIDGenerator *fakeuuid.FakeGenerator
		parser            Parser
		logger            boshlog.Logger
	)

	BeforeEach(func() {
		fakeFs = fakesys.NewFakeFileSystem()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fakeUUIDGenerator = fakeuuid.NewFakeGenerator()
		parser = NewParser(fakeFs, fakeUUIDGenerator, logger)
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

	Context("with a valid manifest", func() {
		BeforeEach(func() {
			contents := `
---
name: fake-deployment-name
cloud_provider:
  template:
    name: fake-cpi-job-name
    release: fake-cpi-release-name
  mbus: http://fake-mbus-user:fake-mbus-password@0.0.0.0:6868
  properties:
    fake-property-name:
      nested-property: fake-property-value
`
			fakeFs.WriteFileString(comboManifestPath, contents)
			fakeFs.ExpandPathExpanded = "/expanded-tmp/fake-ssh-key.pem"
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
				Properties: biproperty.Map{
					"fake-property-name": biproperty.Map{
						"nested-property": "fake-property-value",
					},
				},
				Mbus: "http://fake-mbus-user:fake-mbus-password@0.0.0.0:6868",
			}))
		})
	})

	Context("when ssh tunnel config is present", func() {
		BeforeEach(func() {
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
  mbus: http://fake-mbus-user:fake-mbus-password@0.0.0.0:6868
`
			fakeFs.WriteFileString(comboManifestPath, contents)
			fakeFs.ExpandPathExpanded = "/expanded-tmp/fake-ssh-key.pem"
			fakeUUIDGenerator.GeneratedUUID = "fake-uuid"
		})

		It("generates registry config and populates properties in manifest", func() {
			installationManifest, err := parser.Parse(comboManifestPath)
			Expect(err).ToNot(HaveOccurred())

			Expect(installationManifest).To(Equal(Manifest{
				Name: "fake-deployment-name",
				Template: ReleaseJobRef{
					Name:    "fake-cpi-job-name",
					Release: "fake-cpi-release-name",
				},
				Properties: biproperty.Map{
					"registry": biproperty.Map{
						"host":     "127.0.0.1",
						"port":     6901,
						"username": "registry",
						"password": "fake-uuid",
					},
				},
				Registry: Registry{
					SSHTunnel: SSHTunnel{
						Host:       "54.34.56.8",
						Port:       22,
						User:       "fake-ssh-user",
						PrivateKey: "/expanded-tmp/fake-ssh-key.pem",
					},
					Host:     "127.0.0.1",
					Port:     6901,
					Username: "registry",
					Password: "fake-uuid",
				},
				Mbus: "http://fake-mbus-user:fake-mbus-password@0.0.0.0:6868",
			}))
		})

		Context("when expanding the key file path fails", func() {
			BeforeEach(func() {
				fakeFs.ExpandPathErr = errors.New("fake-expand-error")
			})

			It("uses original path", func() {
				installationManifest, err := parser.Parse(comboManifestPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(installationManifest.Registry.SSHTunnel.PrivateKey).To(Equal("/tmp/fake-ssh-key.pem"))
			})
		})

		Context("when private key is not provided", func() {
			BeforeEach(func() {
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
    password: fake-password
  mbus: http://fake-mbus-user:fake-mbus-password@0.0.0.0:6868
`
				fakeFs.WriteFileString(comboManifestPath, contents)
				fakeFs.ExpandPathExpanded = "/expanded-tmp/fake-ssh-key.pem"
			})

			It("does not expand the path", func() {
				installationManifest, err := parser.Parse(comboManifestPath)
				Expect(err).ToNot(HaveOccurred())

				Expect(installationManifest.Registry.SSHTunnel.PrivateKey).To(Equal(""))
			})
		})
	})
})
