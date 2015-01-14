package release_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	. "github.com/cloudfoundry/bosh-micro-cli/cpi/release"
)

var _ = Describe("Validator", func() {
	var (
		fakeFs *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		fakeFs = fakesys.NewFakeFileSystem()
	})

	It("validates a valid release without error", func() {
		release := bmrel.NewRelease(
			"fake-release-name",
			"fake-release-version",
			[]bmrel.Job{
				{
					Name:        "cpi",
					Fingerprint: "fake-job-1-fingerprint",
					SHA1:        "fake-job-1-sha",
					Templates: map[string]string{
						"cpi.erb": "bin/cpi",
					},
				},
			},
			[]*bmrel.Package{},
			"/some/release/path",
			fakeFs,
		)
		validator := NewValidator()

		err := validator.Validate(release)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when the cpi job is not present", func() {
		var validator Validator
		var release bmrel.Release

		BeforeEach(func() {
			release = bmrel.NewRelease(
				"fake-release-name",
				"fake-release-version",
				[]bmrel.Job{
					{
						Name:        "non-cpi-job",
						Fingerprint: "fake-job-1-fingerprint",
						SHA1:        "fake-job-1-sha",
						Templates: map[string]string{
							"cpi.erb": "bin/cpi",
						},
					},
				},
				[]*bmrel.Package{},
				"/some/release/path",
				fakeFs,
			)
			validator = NewValidator()
		})

		It("returns an error that the cpi job is not present", func() {
			err := validator.Validate(release)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Job 'cpi' is missing from release"))
		})
	})

	Context("when the templates are missing a bin/cpi target", func() {
		var validator Validator
		var release bmrel.Release

		BeforeEach(func() {
			release = bmrel.NewRelease(
				"fake-release-name",
				"fake-release-version",
				[]bmrel.Job{
					{
						Name:        "cpi",
						Fingerprint: "fake-job-1-fingerprint",
						SHA1:        "fake-job-1-sha",
						Templates: map[string]string{
							"cpi.erb": "nonsense",
						},
					},
				},
				[]*bmrel.Package{},
				"/some/release/path",
				fakeFs,
			)
			validator = NewValidator()
		})

		It("returns an error that the bin/cpi template target is missing", func() {
			err := validator.Validate(release)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Job 'cpi' is missing 'bin/cpi' target"))
		})
	})
})
