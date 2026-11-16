package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshreldir "github.com/cloudfoundry/bosh-cli/v7/releasedir"
	fakereldir "github.com/cloudfoundry/bosh-cli/v7/releasedir/releasedirfakes"
)

var _ = Describe("GenerateJobCmd", func() {
	var (
		releaseDir *fakereldir.FakeReleaseDir
		command    cmd.GenerateJobCmd
	)

	BeforeEach(func() {
		releaseDir = &fakereldir.FakeReleaseDir{}
		command = cmd.NewGenerateJobCmd(releaseDir)
	})

	Describe("Run", func() {
		var (
			generateJobOpts opts.GenerateJobOpts
		)

		BeforeEach(func() {
			generateJobOpts = opts.GenerateJobOpts{Args: opts.GenerateJobArgs{Name: "job"}}
		})

		act := func() error { return command.Run(generateJobOpts) }

		It("generates a BPM job by default", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(releaseDir.GenerateJobCallCount()).To(Equal(1))
			Expect(releaseDir.GenerateJobArgsForCall(0)).To(Equal(boshreldir.GenerateJobOptions{
				Name:      "job",
				Templates: map[string]string{},
			}))
		})

		It("passes monit, template and package options through", func() {
			generateJobOpts.Monit = true
			generateJobOpts.Templates = []string{"src/ctl.erb:bin/ctl"}
			generateJobOpts.Packages = []string{"pkg1", "pkg2"}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(releaseDir.GenerateJobArgsForCall(0)).To(Equal(boshreldir.GenerateJobOptions{
				Name:      "job",
				Monit:     true,
				Templates: map[string]string{"src/ctl.erb": "bin/ctl"},
				Packages:  []string{"pkg1", "pkg2"},
			}))
		})

		It("accepts Windows drive-letter source paths", func() {
			generateJobOpts.Templates = []string{`C:\src\ctl.erb:bin/ctl`}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(releaseDir.GenerateJobArgsForCall(0).Templates).To(Equal(
				map[string]string{`C:\src\ctl.erb`: "bin/ctl"}))
		})

		It("returns error if a Windows source path has no destination", func() {
			generateJobOpts.Templates = []string{`C:\src\ctl.erb`}

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected template as SRC:DEST"))
		})

		It("allows colons in the destination", func() {
			generateJobOpts.Templates = []string{"src/ctl.erb:config/foo:bar"}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(releaseDir.GenerateJobArgsForCall(0).Templates).To(Equal(
				map[string]string{"src/ctl.erb": "config/foo:bar"}))
		})

		It("returns error if a template is malformed", func() {
			generateJobOpts.Templates = []string{"src/ctl.erb"}

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected template as SRC:DEST"))
		})

		It("returns error if generating job fails", func() {
			releaseDir.GenerateJobReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
