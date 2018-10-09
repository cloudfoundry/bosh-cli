package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
)

var _ = Describe("DeleteNetworkCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  DeleteNetworkCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = NewDeleteNetworkCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			opts DeleteNetworkOpts
		)

		BeforeEach(func() {
			opts = DeleteNetworkOpts{
				Args: DeleteNetworkArgs{Name: "network-name"},
			}
		})

		act := func() error { return command.Run(opts) }

		It("deletes orphaned network", func() {
			network := &fakedir.FakeOrphanNetwork{}
			director.FindOrphanNetworkReturns(network, nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.FindOrphanNetworkArgsForCall(0)).To(Equal("network-name"))
			Expect(network.DeleteCallCount()).To(Equal(1))
		})

		It("returns error if deleting network failed", func() {
			network := &fakedir.FakeOrphanNetwork{}
			director.FindOrphanNetworkReturns(network, nil)

			network.DeleteReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("does not delete network if confirmation is rejected", func() {
			network := &fakedir.FakeOrphanNetwork{}
			director.FindOrphanNetworkReturns(network, nil)

			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(network.DeleteCallCount()).To(Equal(0))
		})

		It("returns error if finding network failed", func() {
			director.FindOrphanNetworkReturns(nil, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
