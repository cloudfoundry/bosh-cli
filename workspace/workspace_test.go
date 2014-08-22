package workspace_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	. "github.com/cloudfoundry/bosh-micro-cli/workspace"
)

var _ = Describe("Workspace", func() {
	var (
		workspace      Workspace
		fs             *fakesys.FakeFileSystem
		parentDir      string
		logger         boshlog.Logger
		deploymentFile string
		manifestFile   string
		uuid           string
	)
	BeforeEach(func() {
		var err error
		logger = boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		parentDir = "/fake-path"
		deploymentFile = "/fake-path/deployment.json"
		manifestFile = "/fake-path/manifest-file.yml"
		uuid = "abcdef"

		workspace, err = NewWorkspace(fs, parentDir, logger)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("Initialize", func() {
		It("creates blobs directory in .bosh_micro/uuid/blobs", func() {
			err := workspace.Initialize("abcdef")
			Expect(err).ToNot(HaveOccurred())
			Expect(fs.FileExists("/fake-path/.bosh_micro/abcdef/blobs")).To(BeTrue())
		})

		It("returns error if it cannot create blobs dir for this deployment", func() {
			fs.RegisterMkdirAllError("/fake-path/.bosh_micro/abcdef/blobs", errors.New("fake-create-dir"))

			err := workspace.Initialize("abcdef")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Creating blobs dir"))
		})
	})
})
