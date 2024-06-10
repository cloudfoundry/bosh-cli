package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
)

var _ = Describe("TakeSnapshotCmd", func() {
	var (
		deployment *fakedir.FakeDeployment
		command    cmd.TakeSnapshotCmd
	)

	BeforeEach(func() {
		deployment = &fakedir.FakeDeployment{}
		command = cmd.NewTakeSnapshotCmd(deployment)
	})

	Describe("Run", func() {
		var (
			takeSnapshotOpts opts.TakeSnapshotOpts
		)

		BeforeEach(func() {
			takeSnapshotOpts = opts.TakeSnapshotOpts{}
		})

		act := func() error { return command.Run(takeSnapshotOpts) }

		Context("when taking a snapshot of specific instance", func() {
			BeforeEach(func() {
				takeSnapshotOpts.Args.Slug = boshdir.NewInstanceSlug("some-name", "some-id")
			})

			It("take snapshots for a given instance", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.TakeSnapshotCallCount()).To(Equal(1))
				Expect(deployment.TakeSnapshotsCallCount()).To(Equal(0))

				Expect(deployment.TakeSnapshotArgsForCall(0)).To(Equal(
					boshdir.NewInstanceSlug("some-name", "some-id")))
			})

			It("returns error if taking snapshots failed", func() {
				deployment.TakeSnapshotReturns(errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})

		Context("when taking snapshots for the entire deployment", func() {
			It("takes snapshots for the entire deployment", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.TakeSnapshotCallCount()).To(Equal(0))
				Expect(deployment.TakeSnapshotsCallCount()).To(Equal(1))
			})

			It("returns error if taking snapshots failed", func() {
				deployment.TakeSnapshotsReturns(errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})
	})
})
