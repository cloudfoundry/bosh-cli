package validation_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakebmrelease "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
	fakeui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/release/validation"
)

var _ = Describe("Validator", func() {
	var (
		validator         ReleaseValidator
		fakeBoshValidator *fakebmrelease.FakeValidator
		fakeCpiValidator  *fakebmrelease.FakeValidator
		fakeUI            *fakeui.FakeUI
		release           bmrel.Release
	)
	BeforeEach(func() {
		fakeBoshValidator = fakebmrelease.NewFakeValidator()
		fakeCpiValidator = fakebmrelease.NewFakeValidator()
		fakeUI = &fakeui.FakeUI{}
		release = bmrel.Release{TarballPath: "/somepath"}
		validator = NewValidator(fakeBoshValidator, fakeCpiValidator, fakeUI)
	})

	Context("when the release is a valid BOSH CPI release", func() {
		It("returns nil", func() {
			err := validator.Validate(release)
			Expect(err).ToNot(HaveOccurred())
		})

		It("says nothing in the UI", func() {
			Expect(fakeUI.Errors).To(BeEmpty())
			validator.Validate(release)
			Expect(fakeUI.Errors).To(BeEmpty())
		})
	})

	Context("when the release is not a valid BOSH release", func() {
		BeforeEach(func() {
			fakeBoshValidator.ValidateError = errors.New("fake-bosh-error")
		})

		It("returns err", func() {
			err := validator.Validate(release)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-bosh-error"))
		})

		It("errors in the ui", func() {
			validator.Validate(release)
			Expect(fakeUI.Errors).To(ContainElement("CPI release `/somepath' is not a valid BOSH release"))
		})
	})

	Context("when the release is not a valid CPI release", func() {
		BeforeEach(func() {
			fakeCpiValidator.ValidateError = errors.New("fake-cpi-error")
		})

		It("returns err", func() {
			err := validator.Validate(release)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-cpi-error"))
		})

		It("errors in the ui", func() {
			Expect(fakeUI.Errors).To(BeEmpty())
			validator.Validate(release)
			Expect(fakeUI.Errors).To(ContainElement("CPI release `/somepath' is not a valid CPI release"))
		})
	})
})
