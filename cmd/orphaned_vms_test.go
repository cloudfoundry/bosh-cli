package cmd_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/ui/table"
)

var _ = Describe("OrphanedVMsCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  OrphanedVMsCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = NewOrphanedVMsCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			infos []boshdir.OrphanedVM
		)

		Context("when VMs are successfully retrieved", func() {
			BeforeEach(func() {
				infos = []boshdir.OrphanedVM{
					{
						CID:            "a-cid",
						DeploymentName: "deployment",
						InstanceName:   "instance-1/instance-id-1",
						AZName:         "z1",
						IPAddresses:    []string{"127.0.0.10", "127.0.0.11"},
						OrphanedAt:     time.Date(2020, 2, 29, 10, 0, 0, 0, time.UTC),
					},
					{
						CID:            "another-cid",
						DeploymentName: "deployment",
						InstanceName:   "instance-2/instance-id-2",
						AZName:         "z2",
						IPAddresses:    []string{"127.0.0.12"},
						OrphanedAt:     time.Date(2020, 3, 1, 10, 0, 0, 0, time.UTC),
					},
				}

				director.OrphanedVMsReturns(infos, nil)
			})

			It("lists VMs for the deployment", func() {
				Expect(command.Run()).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "orphaned_vms",

					Header: []boshtbl.Header{
						boshtbl.NewHeader("VM CID"),
						boshtbl.NewHeader("Deployment"),
						boshtbl.NewHeader("Instance"),
						boshtbl.NewHeader("AZ"),
						boshtbl.NewHeader("IPs"),
						boshtbl.NewHeader("Orphaned At"),
					},

					SortBy: []boshtbl.ColumnSort{{Column: 5}},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("a-cid"),
							boshtbl.NewValueString("deployment"),
							boshtbl.NewValueString("instance-1/instance-id-1"),
							boshtbl.NewValueString("z1"),
							boshtbl.NewValueStrings([]string{"127.0.0.10", "127.0.0.11"}),
							boshtbl.NewValueTime(time.Date(2020, 2, 29, 10, 0, 0, 0, time.UTC)),
						},
						{
							boshtbl.NewValueString("another-cid"),
							boshtbl.NewValueString("deployment"),
							boshtbl.NewValueString("instance-2/instance-id-2"),
							boshtbl.NewValueString("z2"),
							boshtbl.NewValueStrings([]string{"127.0.0.12"}),
							boshtbl.NewValueTime(time.Date(2020, 3, 1, 10, 0, 0, 0, time.UTC)),
						},
					},
				}))
			})
		})

		Context("when VMs cannot be retrieved", func() {
			BeforeEach(func() {
				director.OrphanedVMsReturns(nil, fmt.Errorf("potato"))
			})

			It("returns an error", func() {
				err := command.Run()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("potato"))
			})
		})
	})
})
