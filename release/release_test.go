package release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/release"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

var _ = Describe("FindJobByName", func() {
	Context("when the job exists", func() {
		var release Release
		var expectedJob bmrel.Job

		BeforeEach(func() {
			expectedJob = bmrel.Job{
				Name: "fake-job-name",
			}
			release = Release{
				Jobs: []bmrel.Job{expectedJob},
			}
		})

		It("returns the job and true", func() {
			actualJob, ok := release.FindJobByName("fake-job-name")
			Expect(actualJob).To(Equal(expectedJob))
			Expect(ok).To(BeTrue())
		})
	})

	Context("when the job does not exist", func() {
		It("returns nil and false", func() {
		})
	})
})
