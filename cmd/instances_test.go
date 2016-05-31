package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	boshdir "github.com/cloudfoundry/bosh-init/director"
	fakedir "github.com/cloudfoundry/bosh-init/director/fakes"
	fakeui "github.com/cloudfoundry/bosh-init/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
)

var _ = Describe("InstancesCmd", func() {
	var (
		ui         *fakeui.FakeUI
		deployment *fakedir.FakeDeployment
		command    InstancesCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		deployment = &fakedir.FakeDeployment{}
		command = NewInstancesCmd(ui, deployment)
	})

	Describe("Run", func() {
		var (
			opts InstancesOpts
		)

		BeforeEach(func() {
			opts = InstancesOpts{}
		})

		act := func() error { return command.Run(opts) }

		Context("when VMs are successfully retrieved", func() {
			var (
				infos          []boshdir.VMInfo
				procCPUTotal   float64
				procMemPercent float64
				procMemKB      uint64
				procUptime     uint64
			)

			BeforeEach(func() {
				index1 := 1
				index2 := 2

				procCPUTotal = 50.40
				procMemPercent = 11.10
				procMemKB = 8000
				procUptime = 349350

				infos = []boshdir.VMInfo{
					{
						JobName:      "job-name",
						Index:        &index1,
						State:        "in1-state",
						ResourcePool: "in1-rp",

						IPs: []string{"in1-ip1", "in1-ip2"},
						DNS: []string{"in1-dns1", "in1-dns2"},

						VMID:               "in1-cid",
						AgentID:            "in1-agent-id",
						ResurrectionPaused: false,

						Vitals: boshdir.VMInfoVitals{
							Load: []string{"0.02", "0.06", "0.11"},

							CPU:  boshdir.VMInfoVitalsCPU{Sys: "0.3", User: "1.2", Wait: "2.1"},
							Mem:  boshdir.VMInfoVitalsMemSize{Percent: "20", KB: "2000"},
							Swap: boshdir.VMInfoVitalsMemSize{Percent: "21", KB: "2100"},

							Disk: map[string]boshdir.VMInfoVitalsDiskSize{
								"system":     boshdir.VMInfoVitalsDiskSize{Percent: "35"},
								"ephemeral":  boshdir.VMInfoVitalsDiskSize{Percent: "45"},
								"persistent": boshdir.VMInfoVitalsDiskSize{Percent: "55"},
							},
						},

						Processes: []boshdir.VMInfoProcess{
							{
								Name:  "in1-proc1-name",
								State: "in1-proc1-state",

								CPU: boshdir.VMInfoVitalsCPU{
									Total: &procCPUTotal,
								},
								Mem: boshdir.VMInfoVitalsMemIntSize{
									Percent: &procMemPercent,
									KB:      &procMemKB,
								},
								Uptime: boshdir.VMInfoVitalsUptime{
									Seconds: &procUptime,
								},
							},
							{
								Name:  "in1-proc2-name",
								State: "in1-proc2-state",
							},
						},
					},
					{
						JobName:      "job-name",
						Index:        &index2,
						State:        "in2-state",
						AZ:           "in2-az",
						ResourcePool: "in2-rp",

						IPs: []string{"in2-ip1"},
						DNS: []string{"in2-dns1"},

						VMID:               "in2-cid",
						AgentID:            "in2-agent-id",
						ResurrectionPaused: true,

						Vitals: boshdir.VMInfoVitals{
							Load: []string{"0.52", "0.56", "0.51"},

							CPU:  boshdir.VMInfoVitalsCPU{Sys: "50.3", User: "51.2", Wait: "52.1"},
							Mem:  boshdir.VMInfoVitalsMemSize{Percent: "60", KB: "6000"},
							Swap: boshdir.VMInfoVitalsMemSize{Percent: "61", KB: "6100"},

							Disk: map[string]boshdir.VMInfoVitalsDiskSize{
								"system":     boshdir.VMInfoVitalsDiskSize{Percent: "75"},
								"ephemeral":  boshdir.VMInfoVitalsDiskSize{Percent: "85"},
								"persistent": boshdir.VMInfoVitalsDiskSize{Percent: "95"},
							},
						},

						Processes: []boshdir.VMInfoProcess{
							{
								Name:  "in2-proc1-name",
								State: "in2-proc1-state",
							},
						},
					},
					{
						JobName:      "",
						Index:        nil,
						State:        "unresponsive agent",
						ResourcePool: "",
					},
				}

				deployment.VMInfosReturns(infos, nil)
			})

			It("lists VMs for the deployment", func() {
				Expect(act()).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "vms",

					HeaderVals: []boshtbl.Value{
						boshtbl.ValueString{"Instance"},
						boshtbl.ValueString{"State"},
						boshtbl.ValueString{"AZ"},
						boshtbl.ValueString{"IPs"},
					},

					SortBy: []boshtbl.ColumnSort{
						{Column: 0, Asc: true},
						{Column: 1, Asc: true},
					},

					Sections: []boshtbl.Section{
						{
							FirstColumn: boshtbl.ValueString{"job-name/1"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"job-name/1"},
									boshtbl.ValueFmt{boshtbl.ValueString{"in1-state"}, true, nil},
									boshtbl.ValueString{},
									boshtbl.ValueStrings{[]string{"in1-ip1", "in1-ip2"}},
								},
							},
						},
						{
							FirstColumn: boshtbl.ValueString{"job-name/2"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"job-name/2"},
									boshtbl.ValueFmt{boshtbl.ValueString{"in2-state"}, true, nil},
									boshtbl.ValueString{"in2-az"},
									boshtbl.ValueStrings{[]string{"in2-ip1"}},
								},
							},
						},
						{
							FirstColumn: boshtbl.ValueString{"?/?"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"?/?"},
									boshtbl.ValueFmt{boshtbl.ValueString{"unresponsive agent"}, true, nil},
									boshtbl.ValueString{},
									boshtbl.ValueStrings{},
								},
							},
						},
					},

					Notes: []string{"(*) Bootstrap node"},
				}))
			})

			It("lists VMs with processes", func() {
				opts.Processes = true

				Expect(act()).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "vms",

					HeaderVals: []boshtbl.Value{
						boshtbl.ValueString{"Instance"},
						boshtbl.ValueString{"Process"},
						boshtbl.ValueString{"State"},
						boshtbl.ValueString{"AZ"},
						boshtbl.ValueString{"IPs"},
					},

					SortBy: []boshtbl.ColumnSort{
						{Column: 0, Asc: true},
						{Column: 1, Asc: true},
					},

					Sections: []boshtbl.Section{
						{
							FirstColumn: boshtbl.ValueString{"job-name/1"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"job-name/1"},
									boshtbl.ValueString{},
									boshtbl.ValueFmt{boshtbl.ValueString{"in1-state"}, true, nil},
									boshtbl.ValueString{},
									boshtbl.ValueStrings{[]string{"in1-ip1", "in1-ip2"}},
								},
								{
									boshtbl.ValueString{},
									boshtbl.ValueString{"in1-proc1-name"},
									boshtbl.ValueFmt{boshtbl.ValueString{"in1-proc1-state"}, true, nil},
									nil,
									nil,
								},
								{
									boshtbl.ValueString{},
									boshtbl.ValueString{"in1-proc2-name"},
									boshtbl.ValueFmt{boshtbl.ValueString{"in1-proc2-state"}, true, nil},
									nil,
									nil,
								},
							},
						},
						{
							FirstColumn: boshtbl.ValueString{"job-name/2"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"job-name/2"},
									boshtbl.ValueString{},
									boshtbl.ValueFmt{boshtbl.ValueString{"in2-state"}, true, nil},
									boshtbl.ValueString{"in2-az"},
									boshtbl.ValueStrings{[]string{"in2-ip1"}},
								},
								{
									boshtbl.ValueString{},
									boshtbl.ValueString{"in2-proc1-name"},
									boshtbl.ValueFmt{boshtbl.ValueString{"in2-proc1-state"}, true, nil},
									nil,
									nil,
								},
							},
						},
						{
							FirstColumn: boshtbl.ValueString{"?/?"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"?/?"},
									boshtbl.ValueString{},
									boshtbl.ValueFmt{boshtbl.ValueString{"unresponsive agent"}, true, nil},
									boshtbl.ValueString{},
									boshtbl.ValueStrings{},
								},
							},
						},
					},

					Notes: []string{"(*) Bootstrap node"},
				}))
			})

			It("lists VMs for the deployment including details", func() {
				opts.Details = true

				Expect(act()).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "vms",

					HeaderVals: []boshtbl.Value{
						boshtbl.ValueString{"Instance"},
						boshtbl.ValueString{"State"},
						boshtbl.ValueString{"AZ"},
						boshtbl.ValueString{"IPs"},
						boshtbl.ValueString{"VM CID"},
						boshtbl.ValueString{"VM Type"},
						boshtbl.ValueString{"Disk CID"},
						boshtbl.ValueString{"Agent ID"},
						boshtbl.ValueString{"Resurrection\nPaused"},
					},

					SortBy: []boshtbl.ColumnSort{
						{Column: 0, Asc: true},
						{Column: 1, Asc: true},
					},

					Sections: []boshtbl.Section{
						{
							FirstColumn: boshtbl.ValueString{"job-name/1"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"job-name/1"},
									boshtbl.ValueFmt{boshtbl.ValueString{"in1-state"}, true, nil},
									boshtbl.ValueString{},
									boshtbl.ValueStrings{[]string{"in1-ip1", "in1-ip2"}},
									boshtbl.ValueString{"in1-cid"},
									boshtbl.ValueString{"in1-rp"},
									boshtbl.ValueString{},
									boshtbl.ValueString{"in1-agent-id"},
									boshtbl.ValueBool{false},
								},
							},
						},
						{
							FirstColumn: boshtbl.ValueString{"job-name/2"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"job-name/2"},
									boshtbl.ValueFmt{boshtbl.ValueString{"in2-state"}, true, nil},
									boshtbl.ValueString{"in2-az"},
									boshtbl.ValueStrings{[]string{"in2-ip1"}},
									boshtbl.ValueString{"in2-cid"},
									boshtbl.ValueString{"in2-rp"},
									boshtbl.ValueString{},
									boshtbl.ValueString{"in2-agent-id"},
									boshtbl.ValueBool{true},
								},
							},
						},
						{
							FirstColumn: boshtbl.ValueString{"?/?"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"?/?"},
									boshtbl.ValueFmt{boshtbl.ValueString{"unresponsive agent"}, true, nil},
									boshtbl.ValueString{},
									boshtbl.ValueStrings{},
									boshtbl.ValueString{},
									boshtbl.ValueString{},
									boshtbl.ValueString{},
									boshtbl.ValueString{},
									boshtbl.ValueBool{false},
								},
							},
						},
					},

					Notes: []string{"(*) Bootstrap node"},
				}))
			})

			It("lists VMs for the deployment including dns", func() {
				opts.DNS = true

				Expect(act()).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "vms",

					HeaderVals: []boshtbl.Value{
						boshtbl.ValueString{"Instance"},
						boshtbl.ValueString{"State"},
						boshtbl.ValueString{"AZ"},
						boshtbl.ValueString{"IPs"},
						boshtbl.ValueString{"DNS A Records"},
					},

					SortBy: []boshtbl.ColumnSort{
						{Column: 0, Asc: true},
						{Column: 1, Asc: true},
					},

					Sections: []boshtbl.Section{
						{
							FirstColumn: boshtbl.ValueString{"job-name/1"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"job-name/1"},
									boshtbl.ValueFmt{boshtbl.ValueString{"in1-state"}, true, nil},
									boshtbl.ValueString{},
									boshtbl.ValueStrings{[]string{"in1-ip1", "in1-ip2"}},
									boshtbl.ValueStrings{[]string{"in1-dns1", "in1-dns2"}},
								},
							},
						},
						{
							FirstColumn: boshtbl.ValueString{"job-name/2"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"job-name/2"},
									boshtbl.ValueFmt{boshtbl.ValueString{"in2-state"}, true, nil},
									boshtbl.ValueString{"in2-az"},
									boshtbl.ValueStrings{[]string{"in2-ip1"}},
									boshtbl.ValueStrings{[]string{"in2-dns1"}},
								},
							},
						},
						{
							FirstColumn: boshtbl.ValueString{"?/?"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"?/?"},
									boshtbl.ValueFmt{boshtbl.ValueString{"unresponsive agent"}, true, nil},
									boshtbl.ValueString{},
									boshtbl.ValueStrings{},
									boshtbl.ValueStrings{},
								},
							},
						},
					},

					Notes: []string{"(*) Bootstrap node"},
				}))
			})

			It("lists VMs for the deployment including vitals and processes", func() {
				opts.Vitals = true
				opts.Processes = true

				Expect(act()).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "vms",

					HeaderVals: []boshtbl.Value{
						boshtbl.ValueString{"Instance"},
						boshtbl.ValueString{"Process"},
						boshtbl.ValueString{"State"},
						boshtbl.ValueString{"AZ"},
						boshtbl.ValueString{"IPs"},
						boshtbl.ValueString{"Uptime"},
						boshtbl.ValueString{"Load\n(1m, 5m, 15m)"},
						boshtbl.ValueString{"CPU\nTotal"},
						boshtbl.ValueString{"CPU\nUser"},
						boshtbl.ValueString{"CPU\nSys"},
						boshtbl.ValueString{"CPU\nWait"},
						boshtbl.ValueString{"Memory\nUsage"},
						boshtbl.ValueString{"Swap\nUsage"},
						boshtbl.ValueString{"System\nDisk Usage"},
						boshtbl.ValueString{"Ephemeral\nDisk Usage"},
						boshtbl.ValueString{"Persistent\nDisk Usage"},
					},

					SortBy: []boshtbl.ColumnSort{
						{Column: 0, Asc: true},
						{Column: 1, Asc: true},
					},

					Sections: []boshtbl.Section{
						{
							FirstColumn: boshtbl.ValueString{"job-name/1"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"job-name/1"},
									boshtbl.ValueString{},
									boshtbl.ValueFmt{boshtbl.ValueString{"in1-state"}, true, nil},
									boshtbl.ValueString{},
									boshtbl.ValueStrings{[]string{"in1-ip1", "in1-ip2"}},
									ValueUptime{},
									boshtbl.ValueString{"0.02, 0.06, 0.11"},
									ValueCPUTotal{},
									NewValueStringPercent("1.2"),
									NewValueStringPercent("0.3"),
									NewValueStringPercent("2.1"),
									ValueMemSize{boshdir.VMInfoVitalsMemSize{Percent: "20", KB: "2000"}},
									ValueMemSize{boshdir.VMInfoVitalsMemSize{Percent: "21", KB: "2100"}},
									ValueDiskSize{boshdir.VMInfoVitalsDiskSize{Percent: "35"}},
									ValueDiskSize{boshdir.VMInfoVitalsDiskSize{Percent: "45"}},
									ValueDiskSize{boshdir.VMInfoVitalsDiskSize{Percent: "55"}},
								},
								{
									boshtbl.ValueString{},
									boshtbl.ValueString{"in1-proc1-name"},
									boshtbl.ValueFmt{boshtbl.ValueString{"in1-proc1-state"}, true, nil},
									nil,
									nil,
									ValueUptime{&procUptime},
									nil,
									ValueCPUTotal{&procCPUTotal},
									nil,
									nil,
									nil,
									ValueMemIntSize{boshdir.VMInfoVitalsMemIntSize{Percent: &procMemPercent, KB: &procMemKB}},
									nil,
									nil,
									nil,
									nil,
								},
								{
									boshtbl.ValueString{},
									boshtbl.ValueString{"in1-proc2-name"},
									boshtbl.ValueFmt{boshtbl.ValueString{"in1-proc2-state"}, true, nil},
									nil,
									nil,
									ValueUptime{},
									nil,
									ValueCPUTotal{},
									nil,
									nil,
									nil,
									ValueMemIntSize{},
									nil,
									nil,
									nil,
									nil,
								},
							},
						},
						{
							FirstColumn: boshtbl.ValueString{"job-name/2"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"job-name/2"},
									boshtbl.ValueString{},
									boshtbl.ValueFmt{boshtbl.ValueString{"in2-state"}, true, nil},
									boshtbl.ValueString{"in2-az"},
									boshtbl.ValueStrings{[]string{"in2-ip1"}},
									ValueUptime{},
									boshtbl.ValueString{"0.52, 0.56, 0.51"},
									ValueCPUTotal{},
									NewValueStringPercent("51.2"),
									NewValueStringPercent("50.3"),
									NewValueStringPercent("52.1"),
									ValueMemSize{boshdir.VMInfoVitalsMemSize{Percent: "60", KB: "6000"}},
									ValueMemSize{boshdir.VMInfoVitalsMemSize{Percent: "61", KB: "6100"}},
									ValueDiskSize{boshdir.VMInfoVitalsDiskSize{Percent: "75"}},
									ValueDiskSize{boshdir.VMInfoVitalsDiskSize{Percent: "85"}},
									ValueDiskSize{boshdir.VMInfoVitalsDiskSize{Percent: "95"}},
								},
								{
									boshtbl.ValueString{},
									boshtbl.ValueString{"in2-proc1-name"},
									boshtbl.ValueFmt{boshtbl.ValueString{"in2-proc1-state"}, true, nil},
									nil,
									nil,
									ValueUptime{},
									nil,
									ValueCPUTotal{},
									nil,
									nil,
									nil,
									ValueMemIntSize{},
									nil,
									nil,
									nil,
									nil,
								},
							},
						},
						{
							FirstColumn: boshtbl.ValueString{"?/?"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"?/?"},
									boshtbl.ValueString{},
									boshtbl.ValueFmt{boshtbl.ValueString{"unresponsive agent"}, true, nil},
									boshtbl.ValueString{},
									boshtbl.ValueStrings{},
									ValueUptime{},
									boshtbl.ValueString{},
									ValueCPUTotal{},
									NewValueStringPercent(""),
									NewValueStringPercent(""),
									NewValueStringPercent(""),
									ValueMemSize{},
									ValueMemSize{},
									ValueDiskSize{},
									ValueDiskSize{},
									ValueDiskSize{},
								},
							},
						},
					},

					Notes: []string{"(*) Bootstrap node"},
				}))
			})

			It("lists failing (non-running) VMs", func() {
				opts.Failing = true

				// Hides second VM
				infos[1].State = "running"
				infos[1].Processes[0].State = "running"

				Expect(act()).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "vms",

					HeaderVals: []boshtbl.Value{
						boshtbl.ValueString{"Instance"},
						boshtbl.ValueString{"State"},
						boshtbl.ValueString{"AZ"},
						boshtbl.ValueString{"IPs"},
					},

					SortBy: []boshtbl.ColumnSort{
						{Column: 0, Asc: true},
						{Column: 1, Asc: true},
					},

					Sections: []boshtbl.Section{
						{
							FirstColumn: boshtbl.ValueString{"job-name/1"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"job-name/1"},
									boshtbl.ValueFmt{boshtbl.ValueString{"in1-state"}, true, nil},
									boshtbl.ValueString{},
									boshtbl.ValueStrings{[]string{"in1-ip1", "in1-ip2"}},
								},
							},
						},
						{
							FirstColumn: boshtbl.ValueString{"?/?"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"?/?"},
									boshtbl.ValueFmt{boshtbl.ValueString{"unresponsive agent"}, true, nil},
									boshtbl.ValueString{},
									boshtbl.ValueStrings{},
								},
							},
						},
					},

					Notes: []string{"(*) Bootstrap node"},
				}))
			})

			It("includes failing processes when listing failing (non-running) VMs and processes", func() {
				opts.Failing = true
				opts.Processes = true

				// Hides first process in the first VM
				infos[0].Processes[0].State = "running"

				// Hides second VM completely
				infos[1].State = "running"
				infos[1].Processes[0].State = "running"

				Expect(act()).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Content: "vms",

					HeaderVals: []boshtbl.Value{
						boshtbl.ValueString{"Instance"},
						boshtbl.ValueString{"Process"},
						boshtbl.ValueString{"State"},
						boshtbl.ValueString{"AZ"},
						boshtbl.ValueString{"IPs"},
					},

					SortBy: []boshtbl.ColumnSort{
						{Column: 0, Asc: true},
						{Column: 1, Asc: true},
					},

					Sections: []boshtbl.Section{
						{
							FirstColumn: boshtbl.ValueString{"job-name/1"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"job-name/1"},
									boshtbl.ValueString{},
									boshtbl.ValueFmt{boshtbl.ValueString{"in1-state"}, true, nil},
									boshtbl.ValueString{},
									boshtbl.ValueStrings{[]string{"in1-ip1", "in1-ip2"}},
								},
								{
									boshtbl.ValueString{},
									boshtbl.ValueString{"in1-proc2-name"},
									boshtbl.ValueFmt{boshtbl.ValueString{"in1-proc2-state"}, true, nil},
									nil,
									nil,
								},
							},
						},
						{
							FirstColumn: boshtbl.ValueString{"?/?"},
							Rows: [][]boshtbl.Value{
								{
									boshtbl.ValueString{"?/?"},
									boshtbl.ValueString{},
									boshtbl.ValueFmt{boshtbl.ValueString{"unresponsive agent"}, true, nil},
									boshtbl.ValueString{},
									boshtbl.ValueStrings{},
								},
							},
						},
					},

					Notes: []string{"(*) Bootstrap node"},
				}))
			})
		})

		It("returns error if VMs cannot be retrieved", func() {
			deployment.VMInfosReturns(nil, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
