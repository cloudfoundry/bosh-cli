package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("DeleteReleaseCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  cmd.DeleteReleaseCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewDeleteReleaseCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			deleteReleaseOpts opts.DeleteReleaseOpts
		)

		BeforeEach(func() {
			deleteReleaseOpts = opts.DeleteReleaseOpts{}
		})

		act := func() error { return command.Run(deleteReleaseOpts) }

		Context("when release series is requested for deletion", func() {
			var (
				releaseSeries *fakedir.FakeReleaseSeries
			)

			BeforeEach(func() {
				deleteReleaseOpts.Args.Slug = boshdir.NewReleaseOrSeriesSlug("some-name", "")

				releaseSeries = &fakedir.FakeReleaseSeries{}
				director.FindReleaseSeriesReturns(releaseSeries, nil)
				releaseSeries.ExistsReturns(true, nil)
			})

			It("deletes release series that exists", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.FindReleaseSeriesCallCount()).To(Equal(1))
				Expect(director.FindReleaseSeriesArgsForCall(0)).To(Equal(
					boshdir.NewReleaseSeriesSlug("some-name")))

				Expect(releaseSeries.ExistsCallCount()).To(Equal(1))
				Expect(releaseSeries.DeleteCallCount()).To(Equal(1))
				Expect(releaseSeries.DeleteArgsForCall(0)).To(BeFalse())
			})

			It("does not delete release series which does not exist", func() {
				deleteReleaseOpts.Args.Slug = boshdir.NewReleaseOrSeriesSlug("not-existing-release", "")
				releaseSeries.ExistsReturns(false, nil)
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.FindReleaseSeriesCallCount()).To(Equal(1))
				Expect(director.FindReleaseSeriesArgsForCall(0)).To(Equal(
					boshdir.NewReleaseSeriesSlug("not-existing-release")))

				Expect(releaseSeries.ExistsCallCount()).To(Equal(1))
				Expect(releaseSeries.DeleteCallCount()).To(Equal(0))
			})

			It("deletes release series forcefully if requested", func() {
				deleteReleaseOpts.Force = true

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(releaseSeries.DeleteCallCount()).To(Equal(1))
				Expect(releaseSeries.ExistsCallCount()).To(Equal(1))
				Expect(releaseSeries.DeleteArgsForCall(0)).To(BeTrue())
			})

			It("does not delete release series if confirmation is rejected", func() {
				ui.AskedConfirmationErr = errors.New("stop")

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("stop"))

				Expect(releaseSeries.DeleteCallCount()).To(Equal(0))
			})

			It("returns error if deleting release series failed", func() {
				releaseSeries.DeleteReturns(errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})

			It("returns error if finding release series failed", func() {
				director.FindReleaseSeriesReturns(nil, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))

				Expect(releaseSeries.DeleteCallCount()).To(Equal(0))
			})

			It("returns error if release series existence check failed", func() {
				releaseSeries.ExistsReturns(false, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))

				Expect(releaseSeries.DeleteCallCount()).To(Equal(0))
			})
		})

		Context("when release (not series) is requested for deletion", func() {
			var (
				release *fakedir.FakeRelease
			)

			BeforeEach(func() {
				deleteReleaseOpts.Args.Slug = boshdir.NewReleaseOrSeriesSlug("some-name", "some-version")

				release = &fakedir.FakeRelease{}
				director.FindReleaseReturns(release, nil)
				release.ExistsReturns(true, nil)
			})

			It("deletes release that exists", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.FindReleaseCallCount()).To(Equal(1))
				Expect(director.FindReleaseArgsForCall(0)).To(Equal(
					boshdir.NewReleaseSlug("some-name", "some-version")))

				Expect(release.ExistsCallCount()).To(Equal(1))
				Expect(release.DeleteCallCount()).To(Equal(1))
				Expect(release.DeleteArgsForCall(0)).To(BeFalse())
			})

			It("does not delete release which does not exist", func() {
				deleteReleaseOpts.Args.Slug = boshdir.NewReleaseOrSeriesSlug("some-other-name", "some-version")
				release.ExistsReturns(false, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(director.FindReleaseCallCount()).To(Equal(1))
				Expect(director.FindReleaseArgsForCall(0)).To(Equal(
					boshdir.NewReleaseSlug("some-other-name", "some-version")))

				Expect(release.ExistsCallCount()).To(Equal(1))
				Expect(release.DeleteCallCount()).To(Equal(0))
			})

			It("deletes release forcefully if requested", func() {
				deleteReleaseOpts.Force = true

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(release.ExistsCallCount()).To(Equal(1))
				Expect(release.DeleteCallCount()).To(Equal(1))
				Expect(release.DeleteArgsForCall(0)).To(BeTrue())
			})

			It("does not delete release if confirmation is rejected", func() {
				ui.AskedConfirmationErr = errors.New("stop")

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("stop"))

				Expect(release.DeleteCallCount()).To(Equal(0))
			})

			It("returns error if deleting release failed", func() {
				release.DeleteReturns(errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})

			It("returns error if release existence check failed", func() {
				release.ExistsReturns(false, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
				Expect(release.DeleteCallCount()).To(Equal(0))
			})

			It("returns error if finding release failed", func() {
				director.FindReleaseReturns(nil, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))

				Expect(release.DeleteCallCount()).To(Equal(0))
			})
		})
	})
})
