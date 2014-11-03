package validation_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	. "github.com/cloudfoundry/bosh-micro-cli/release/validation"
)

var _ = Describe("CpiValidator", func() {
	It("validates a valid release without error", func() {
		release := bmrel.Release{
			Jobs: []bmrel.Job{
				{
					Name:        "cpi",
					Fingerprint: "fake-job-1-fingerprint",
					SHA1:        "fake-job-1-sha",
					Templates: map[string]string{
						"cpi.erb": "bin/cpi",
					},
				},
			},
		}
		validator := NewCpiValidator()

		err := validator.Validate(release)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when the cpi job is not present", func() {
		var validator CpiValidator
		var release bmrel.Release

		BeforeEach(func() {
			release = bmrel.Release{
				Jobs: []bmrel.Job{
					{
						Name:        "non-cpi-job",
						Fingerprint: "fake-job-1-fingerprint",
						SHA1:        "fake-job-1-sha",
						Templates: map[string]string{
							"cpi.erb": "bin/cpi",
						},
					},
				},
			}
			validator = NewCpiValidator()
		})

		It("returns an error that the cpi job is not present", func() {
			err := validator.Validate(release)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Job `cpi' is missing from release"))
		})
	})

	Context("when the templates are missing a bin/cpi target", func() {
		var validator CpiValidator
		var release bmrel.Release

		BeforeEach(func() {
			release = bmrel.Release{
				Jobs: []bmrel.Job{
					{
						Name:        "cpi",
						Fingerprint: "fake-job-1-fingerprint",
						SHA1:        "fake-job-1-sha",
						Templates: map[string]string{
							"cpi.erb": "nonsense",
						},
					},
				},
			}
			validator = NewCpiValidator()
		})

		It("returns an error that the bin/cpi template target is missing", func() {
			err := validator.Validate(release)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Job `cpi' is missing bin/cpi target"))
		})
	})
})
