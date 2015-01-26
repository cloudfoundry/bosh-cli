package release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment/release"

	"code.google.com/p/gomock/gomock"
	mock_release_set "github.com/cloudfoundry/bosh-micro-cli/release/set/mocks"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	bmdeplmanifest "github.com/cloudfoundry/bosh-micro-cli/deployment/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fake_release "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
)

var _ = Describe("JobResolver", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	var (
		mockReleaseSetResolver *mock_release_set.MockResolver
		fakeRelease            *fake_release.FakeRelease

		fakeReleaseJob0 bmrel.Job
		fakeReleaseJob1 bmrel.Job

		jobResolver JobResolver
	)

	BeforeEach(func() {
		mockReleaseSetResolver = mock_release_set.NewMockResolver(mockCtrl)

		fakeRelease = fake_release.New("fake-release-name", "fake-release-version")

		fakeReleaseJob0 = bmrel.Job{
			Name:        "fake-release-job-name-0",
			Fingerprint: "fake-release-job-fingerprint-0",
		}
		fakeReleaseJob1 = bmrel.Job{
			Name:        "fake-release-job-name-1",
			Fingerprint: "fake-release-job-fingerprint-1",
		}
	})

	JustBeforeEach(func() {
		jobResolver = NewJobResolver(mockReleaseSetResolver)

		fakeRelease.ReleaseJobs = []bmrel.Job{fakeReleaseJob0, fakeReleaseJob1}
	})

	Describe("Resolve", func() {
		It("Returns the matching release job", func() {
			mockReleaseSetResolver.EXPECT().Find("fake-release-name").Return(fakeRelease, nil)

			releaseJob, err := jobResolver.Resolve(bmdeplmanifest.ReleaseJobRef{
				Name:    "fake-release-job-name-0",
				Release: "fake-release-name",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(releaseJob).To(Equal(fakeReleaseJob0))
		})

		It("Returns an error, when the job is not in the release", func() {
			mockReleaseSetResolver.EXPECT().Find("fake-release-name").Return(fakeRelease, nil)

			_, err := jobResolver.Resolve(bmdeplmanifest.ReleaseJobRef{
				Name:    "fake-missing-release-job-name",
				Release: "fake-release-name",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Finding job 'fake-missing-release-job-name' in release 'fake-release-name'"))
		})

		It("Returns an error, when the release is not in resolvable", func() {
			mockReleaseSetResolver.EXPECT().Find("fake-missing-release-name").Return(nil, bosherr.Error("fake-release-resolver-find-error"))

			_, err := jobResolver.Resolve(bmdeplmanifest.ReleaseJobRef{
				Name:    "fake-release-job-name-0",
				Release: "fake-missing-release-name",
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Resolving release 'fake-missing-release-name'"))
			Expect(err.Error()).To(ContainSubstring("fake-release-resolver-find-error"))
		})
	})

	Describe("ResolveEach", func() {
		It("Returns the matching release jobs", func() {
			mockReleaseSetResolver.EXPECT().Find("fake-release-name").Return(fakeRelease, nil).Times(2)

			releaseJobs, err := jobResolver.ResolveEach([]bmdeplmanifest.ReleaseJobRef{
				{Name: "fake-release-job-name-0", Release: "fake-release-name"},
				{Name: "fake-release-job-name-1", Release: "fake-release-name"},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(releaseJobs).To(Equal([]bmrel.Job{fakeReleaseJob0, fakeReleaseJob1}))
		})

		It("Returns an error, when one of the jobs is not in the release", func() {
			mockReleaseSetResolver.EXPECT().Find("fake-release-name").Return(fakeRelease, nil)

			_, err := jobResolver.ResolveEach([]bmdeplmanifest.ReleaseJobRef{
				{Name: "fake-missing-release-job-name", Release: "fake-release-name"},
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Finding job 'fake-missing-release-job-name' in release 'fake-release-name'"))
		})

		It("Returns an error, when one of the releases is not in resolvable", func() {
			mockReleaseSetResolver.EXPECT().Find("fake-missing-release-name").Return(nil, bosherr.Error("fake-release-resolver-find-error"))

			_, err := jobResolver.ResolveEach([]bmdeplmanifest.ReleaseJobRef{
				{Name: "fake-release-job-name-0", Release: "fake-missing-release-name"},
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Resolving release 'fake-missing-release-name'"))
			Expect(err.Error()).To(ContainSubstring("fake-release-resolver-find-error"))
		})
	})
})
