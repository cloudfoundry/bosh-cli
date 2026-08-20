package release_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/deployment/release"
	bireljob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	fakerel "github.com/cloudfoundry/bosh-cli/v7/release/releasefakes"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
)

var _ = Describe("JobResolver", func() {
	var (
		mockReleaseManager *fakerel.FakeManager
		release            *fakerel.FakeRelease
		jobResolver        JobResolver
	)

	BeforeEach(func() {
		mockReleaseManager = &fakerel.FakeManager{}

		release = &fakerel.FakeRelease{
			NameStub:    func() string { return "rel-name" },
			VersionStub: func() string { return "rel-ver" },
		}

		jobResolver = NewJobResolver(mockReleaseManager)
	})

	Describe("Resolve", func() {
		It("Returns the matching release job", func() {
			job0 := bireljob.NewJob(NewResource("job0", "job0-fp", nil))
			mockReleaseManager.FindReturns(release, true)
			release.FindJobByNameStub = func(name string) (bireljob.Job, bool) {
				Expect(name).To(Equal("job0"))
				return *job0, true
			}

			releaseJob, err := jobResolver.Resolve("job0", "rel-name")
			Expect(err).ToNot(HaveOccurred())
			Expect(releaseJob).To(Equal(*job0))

			Expect(mockReleaseManager.FindCallCount()).To(Equal(1))
			Expect(mockReleaseManager.FindArgsForCall(0)).To(Equal("rel-name"))
		})

		It("Returns an error, when the job is not in the release", func() {
			mockReleaseManager.FindReturns(release, true)
			release.FindJobByNameReturns(bireljob.Job{}, false)

			_, err := jobResolver.Resolve("fake-missing-release-job-name", "rel-name")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Finding job 'fake-missing-release-job-name' in release 'rel-name'"))
		})

		It("Returns an error, when the release is not in resolvable", func() {
			mockReleaseManager.FindReturns(nil, false)

			_, err := jobResolver.Resolve("job0", "fake-missing-release-name")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Finding release 'fake-missing-release-name'"))
		})
	})
})
