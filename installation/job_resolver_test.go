package installation_test

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/deployment/release/releasefakes"
	"github.com/cloudfoundry/bosh-cli/v7/installation"
	biinstallmanifest "github.com/cloudfoundry/bosh-cli/v7/installation/manifest"
	bireljob "github.com/cloudfoundry/bosh-cli/v7/release/job"
	. "github.com/cloudfoundry/bosh-cli/v7/release/resource"
)

var _ = Describe("JobResolver", func() {
	var (
		mockReleaseJobResolver *releasefakes.FakeJobResolver
		resolver               installation.JobResolver
		releaseJob             bireljob.Job
		manifest               biinstallmanifest.Manifest
	)

	BeforeEach(func() {
		mockReleaseJobResolver = &releasefakes.FakeJobResolver{}

		manifest = biinstallmanifest.Manifest{
			Name: "fake-installation-name",
			Templates: []biinstallmanifest.ReleaseJobRef{
				{Name: "fake-cpi-job-name", Release: "fake-cpi-release-name"},
			},
			Properties: biproperty.Map{
				"fake-installation-property": "fake-installation-property-value",
			},
		}

		job := bireljob.NewJob(NewResource("cpi", "fake-release-job-fingerprint", nil))
		releaseJob = *job
	})

	JustBeforeEach(func() {
		resolver = installation.NewJobResolver(mockReleaseJobResolver)
		mockReleaseJobResolver.ResolveReturns(releaseJob, nil)
	})

	Describe("From", func() {
		It("when the release does contain a 'cpi' job returns release jobs", func() {
			jobs, err := resolver.From(manifest)
			Expect(err).ToNot(HaveOccurred())
			Expect(jobs).To(Equal([]bireljob.Job{releaseJob}))

			jobName, releaseName := mockReleaseJobResolver.ResolveArgsForCall(0)
			Expect(jobName).To(Equal("fake-cpi-job-name"))
			Expect(releaseName).To(Equal("fake-cpi-release-name"))
		})

		It("when the release does not contain a 'cpi' job returns an error", func() {
			mockReleaseJobResolver.ResolveReturns(bireljob.Job{}, bosherr.Error("fake-job-resolve-error"))
			_, err := resolver.From(manifest)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-job-resolve-error"))
		})
	})
})
