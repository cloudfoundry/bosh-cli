package release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"

	. "github.com/cloudfoundry/bosh-micro-cli/release"
)

var _ = Describe("Release", func() {
	var (
		release     Release
		expectedJob bmreljob.Job
		fakeFS      *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		expectedJob = bmreljob.Job{
			Name: "fake-job-name",
		}
		fakeFS = fakesys.NewFakeFileSystem()
		release = NewRelease(
			"fake-release-name",
			"fake-release-version",
			[]bmreljob.Job{expectedJob},
			[]*bmrelpkg.Package{},
			"fake-extracted-path",
			fakeFS,
		)
	})

	Describe("FindJobByName", func() {
		Context("when the job exists", func() {
			It("returns the job and true", func() {
				actualJob, ok := release.FindJobByName("fake-job-name")
				Expect(actualJob).To(Equal(expectedJob))
				Expect(ok).To(BeTrue())
			})
		})

		Context("when the job does not exist", func() {
			It("returns nil and false", func() {
				_, ok := release.FindJobByName("fake-non-existent-job")
				Expect(ok).To(BeFalse())
			})
		})
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			fakeFS.WriteFileString("fake-extracted-path", "")
		})

		It("deletes the extracted release path", func() {
			Expect(fakeFS.FileExists("fake-extracted-path")).To(BeTrue())
			err := release.Delete()
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeFS.FileExists("fake-extracted-path")).To(BeFalse())
		})
	})

	Describe("Exists", func() {
		BeforeEach(func() {
			fakeFS.WriteFileString("fake-extracted-path", "")
		})

		It("returns false after deletion", func() {
			Expect(release.Exists()).To(BeTrue())
			err := release.Delete()
			Expect(err).ToNot(HaveOccurred())
			Expect(release.Exists()).To(BeFalse())
		})
	})
})
