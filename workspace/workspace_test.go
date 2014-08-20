package workspace_test

import (
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"

	. "github.com/cloudfoundry/bosh-micro-cli/workspace"
)

var _ = Describe("Workspace", func() {
	var (
		fs        *fakesys.FakeFileSystem
		uuidGen   *fakeuuid.FakeGenerator
		config    bmconfig.Config
		parentDir string
	)

	Context("Initialize", func() {
		var deploymentFilePath string
		BeforeEach(func() {
			fs = fakesys.NewFakeFileSystem()
			uuidGen = &fakeuuid.FakeGenerator{}
			parentDir = "/fake-path"
			uuidGen.GeneratedUuid = "abcdef"
			deploymentFilePath = "/fake-path/deployment.json"
		})

		Context("when the deployment is set in the config", func() {
			BeforeEach(func() {
				config = bmconfig.Config{
					Deployment: "/fake-path/manifest.yml",
				}
			})

			It("creates deployment.json", func() {
				_, err := NewWorkspace(fs, config, uuidGen, parentDir)
				Expect(err).ToNot(HaveOccurred())
				Expect(fs.FileExists(deploymentFilePath)).To(BeTrue())
			})

			It("stores a UUID in deploymenet.json", func() {
				uuidGen.GeneratedUuid = "abcdef"
				_, err := NewWorkspace(fs, config, uuidGen, parentDir)
				Expect(err).ToNot(HaveOccurred())
				deploymentContent, err := fs.ReadFile(deploymentFilePath)
				Expect(err).ToNot(HaveOccurred())

				deploymentFile := DeploymentFile{}
				err = json.Unmarshal(deploymentContent, &deploymentFile)
				Expect(err).ToNot(HaveOccurred())
				Expect(deploymentFile.UUID).ToNot(BeEmpty())
			})

			It("creates blobs directory in .bosh_micro/uuid/blobs where uuid is from config", func() {
				uuidGen.GeneratedUuid = "abcdef"
				_, err := NewWorkspace(fs, config, uuidGen, parentDir)
				Expect(err).ToNot(HaveOccurred())
				Expect(fs.FileExists("/fake-path/.bosh_micro/abcdef/blobs")).To(BeTrue())
			})

			It("returns error when it cannot generate a UUID", func() {
				uuidGen.GenerateError = errors.New("fake-generate-error")
				_, err := NewWorkspace(fs, config, uuidGen, parentDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Generating UUID"))
			})

			It("returns error if it cannot create blobs dir for this deployment", func() {
				fs.RegisterMkdirAllError("/fake-path/.bosh_micro/abcdef/blobs", errors.New("fake-create-dir"))
				uuidGen.GeneratedUuid = "abcdef"

				_, err := NewWorkspace(fs, config, uuidGen, parentDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Creating blobs dir"))
			})

			It("errors if it cannot create deployment.json", func() {
				fs.WriteToFileError = errors.New("fake-write-file")
				_, err := NewWorkspace(fs, config, uuidGen, parentDir)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Writing deployment file"))
			})
		})

		Context("when the deployment is set in the config and deployment.json exists", func() {
			BeforeEach(func() {
				config = bmconfig.Config{
					Deployment: "/fake-path/manifest.yml",
				}
				deploymentJSON, err := json.MarshalIndent(DeploymentFile{UUID: "existing-uuid"}, "", " ")
				Expect(err).ToNot(HaveOccurred())
				fs.WriteFileString("/fake-path/deployment.json", string(deploymentJSON))
			})

			It("does not create a blobstore directory", func() {
				uuidGen.GeneratedUuid = "abcdef"
				_, err := NewWorkspace(fs, config, uuidGen, parentDir)
				Expect(err).ToNot(HaveOccurred())

				Expect(fs.FileExists("/fake-path/.bosh_micro/abcdef/blobs")).To(BeFalse())
			})
		})

		Context("when there is no deployment", func() {
			BeforeEach(func() {
				config = bmconfig.Config{}
			})

			It("does not create a deployment file", func() {
				uuidGen.GeneratedUuid = "abcdef"
				_, err := NewWorkspace(fs, config, uuidGen, parentDir)
				Expect(err).ToNot(HaveOccurred())

				Expect(fs.FileExists("deployment.json")).To(BeFalse())
				Expect(fs.FileExists("/fake-path/deployment.json")).To(BeFalse())
			})

			It("does not create a blobstore directory", func() {
				uuidGen.GeneratedUuid = "abcdef"
				_, err := NewWorkspace(fs, config, uuidGen, parentDir)
				Expect(err).ToNot(HaveOccurred())

				Expect(fs.FileExists("/fake-path/.bosh_micro/abcdef/blobs")).To(BeFalse())
			})
		})

		Describe("MicroBoshPath", func() {
			Context("when there is a deployment", func() {
				BeforeEach(func() {
					config = bmconfig.Config{
						Deployment: "/fake-path/manifest.yml",
					}
					uuidGen.GeneratedUuid = "fake-uuid"
				})
				It("returns a new bosh micro path", func() {
					workspace, err := NewWorkspace(fs, config, uuidGen, parentDir)
					Expect(err).ToNot(HaveOccurred())

					Expect(workspace.MicroBoshPath()).To(Equal("/fake-path/.bosh_micro/fake-uuid"))
				})
			})

			Context("when there is a deployment and a uuid", func() {
				BeforeEach(func() {
					config = bmconfig.Config{
						Deployment: "/fake-path/manifest.yml",
					}
					uuidGen.GeneratedUuid = "fake-uuid"

				})

				It("returns the bosh micro path of the deployment", func() {
					deploymentJSON, err := json.MarshalIndent(DeploymentFile{UUID: "existing-uuid"}, "", " ")
					Expect(err).ToNot(HaveOccurred())
					fs.WriteFileString("/fake-path/deployment.json", string(deploymentJSON))

					workspace, err := NewWorkspace(fs, config, uuidGen, parentDir)
					Expect(err).ToNot(HaveOccurred())

					Expect(workspace.MicroBoshPath()).To(Equal("/fake-path/.bosh_micro/existing-uuid"))
				})

				It("errors when it cannot read the deployment file", func() {
					fs.WriteFileString("/fake-path/deployment.json", "")
					fs.ReadFileError = errors.New("fake-read-error")

					_, err := NewWorkspace(fs, config, uuidGen, parentDir)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Reading deployment file"))
				})

				It("errors when it cannot unmarshal the deployment file", func() {
					fs.WriteFileString("/fake-path/deployment.json", "---invalid json---")

					_, err := NewWorkspace(fs, config, uuidGen, parentDir)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Unmarshalling deployment file"))
				})
			})
		})
	})
})
