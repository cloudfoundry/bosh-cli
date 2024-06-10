package cmd_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("NetworksCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  cmd.NetworksCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewNetworksCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			networksOpts opts.NetworksOpts
		)

		BeforeEach(func() {
			networksOpts = opts.NetworksOpts{}
		})

		act := func() error { return command.Run(networksOpts) }

		Context("when orphaned networks requested", func() {
			BeforeEach(func() {
				networksOpts.Orphaned = true
			})

			It("lists networks", func() {
				networks := []boshdir.OrphanNetwork{
					&fakedir.FakeOrphanNetwork{
						NameStub: func() string { return "fake-network" },
						TypeStub: func() string { return "manual" },
						CreatedAtStub: func() time.Time {
							return time.Date(2009, time.March, 10, 23, 0, 0, 0, time.UTC)
						},
						OrphanedAtStub: func() time.Time {
							return time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
						},
					},
				}

				director.OrphanNetworksReturns(networks, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table.Rows[0][0]).To(ContainSubstring("fake-network"))
			})

			It("returns error if orphaned networks cannot be retrieved", func() {
				director.OrphanNetworksReturns(nil, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})

		It("returns error if orphaned disks were not requested", func() {
			Expect(act()).To(Equal(errors.New("Only --orphaned is supported")))
		})
	})
})
