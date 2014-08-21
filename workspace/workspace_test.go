package workspace_test

import (
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	. "github.com/cloudfoundry/bosh-micro-cli/workspace"
)

var _ = Describe("Workspace", func() {
	var (
		workspace          Workspace
		fs                 *fakesys.FakeFileSystem
		uuidGen            *fakeuuid.FakeGenerator
		parentDir          string
		logger             boshlog.Logger
		deploymentFilePath string
		manifestFile       string
	)
	BeforeEach(func() {
		var err error
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		uuidGen = &fakeuuid.FakeGenerator{}
		parentDir = "/fake-path"
		uuidGen.GeneratedUuid = "abcdef"
		deploymentFilePath = "/fake-path/deployment.json"
		manifestFile = "/fake-path/manifest-file.yml"

		workspace, err = NewWorkspace(fs, uuidGen, parentDir, logger)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("Initialize", func() {
		It("creates deployment.json", func() {
			err := workspace.Initialize(manifestFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(fs.FileExists(deploymentFilePath)).To(BeTrue())
		})

		It("stores a UUID in deployment.json", func() {
			err := workspace.Initialize(manifestFile)
			Expect(err).ToNot(HaveOccurred())
			deploymentContent, err := fs.ReadFile(deploymentFilePath)
			Expect(err).ToNot(HaveOccurred())

			deploymentFile := DeploymentFile{}
			err = json.Unmarshal(deploymentContent, &deploymentFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(deploymentFile.UUID).ToNot(BeEmpty())
		})

		It("creates blobs directory in .bosh_micro/uuid/blobs", func() {
			uuidGen.GeneratedUuid = "abcdef"
			err := workspace.Initialize(manifestFile)
			Expect(err).ToNot(HaveOccurred())
			Expect(fs.FileExists("/fake-path/.bosh_micro/abcdef/blobs")).To(BeTrue())
		})

		It("returns error when it cannot generate a UUID", func() {
			uuidGen.GenerateError = errors.New("fake-generate-error")
			err := workspace.Initialize(manifestFile)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Generating UUID"))
		})

		It("returns error if it cannot create blobs dir for this deployment", func() {
			fs.RegisterMkdirAllError("/fake-path/.bosh_micro/abcdef/blobs", errors.New("fake-create-dir"))
			uuidGen.GeneratedUuid = "abcdef"

			err := workspace.Initialize(manifestFile)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Creating blobs dir"))
		})

		It("errors if it cannot create deployment.json", func() {
			fs.WriteToFileError = errors.New("fake-write-file")
			err := workspace.Initialize(manifestFile)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Writing deployment file"))
		})
	})

	Describe("Load", func() {
		It("errors when it cannot read the deployment file", func() {
			fs.WriteFileString("/fake-path/deployment.json", "")
			fs.ReadFileError = errors.New("fake-read-error")

			err := workspace.Load(manifestFile)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Reading deployment file"))
		})

		It("errors when it cannot unmarshal the deployment file", func() {
			fs.WriteFileString("/fake-path/deployment.json", "---invalid json---")

			err := workspace.Load(manifestFile)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unmarshalling deployment file"))
		})
	})

	Describe("MicroBoshPath", func() {
		Context("after initializing", func() {
			It("returns a new bosh micro path", func() {
				uuidGen.GeneratedUuid = "new-uuid"
				err := workspace.Initialize(manifestFile)
				Expect(err).ToNot(HaveOccurred())

				Expect(workspace.MicroBoshPath()).To(Equal("/fake-path/.bosh_micro/new-uuid"))
			})
		})

		Context("after loading", func() {
			It("loads existing deployment uuid", func() {
				deploymentContent, err := json.Marshal(DeploymentFile{UUID: "loaded-uuid"})
				Expect(err).ToNot(HaveOccurred())

				fs.WriteFile("/fake-path/deployment.json", deploymentContent)
				err = workspace.Load(manifestFile)
				Expect(err).ToNot(HaveOccurred())

				Expect(workspace.MicroBoshPath()).To(Equal("/fake-path/.bosh_micro/loaded-uuid"))
			})
		})
	})
})
