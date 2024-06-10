package cmd_test

import (
	"errors"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

var _ = Describe("InstancesCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  cmd.InstancesCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewInstancesCmd(ui, director, 1)
	})

	Describe("Run", func() {
		var (
			instancesOpts  opts.InstancesOpts
			infos          []boshdir.VMInfo
			procCPUTotal   float64
			procMemPercent float64
			procMemKB      uint64
			procUptime     uint64
		)

		act := func() error { return command.Run(instancesOpts) }

		BeforeEach(func() {
			instancesOpts = opts.InstancesOpts{}

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
					ProcessState: "in1-process-state",
					ResourcePool: "in1-rp",

					IPs:        []string{"in1-ip1", "in1-ip2"},
					Deployment: "dep",

					State:       "in1-state",
					VMID:        "in1-cid",
					AgentID:     "in1-agent-id",
					Ignore:      true,
					DiskIDs:     []string{"diskcid1", "diskcid2"},
					VMCreatedAt: time.Date(2016, time.January, 9, 6, 23, 25, 0, time.UTC),

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
					ProcessState: "in2-process-state",
					AZ:           "in2-az",
					ResourcePool: "in2-rp",

					IPs:        []string{"in2-ip1"},
					Deployment: "dep",

					State:       "in2-state",
					VMID:        "in2-cid",
					AgentID:     "in2-agent-id",
					Ignore:      false,
					DiskIDs:     []string{"diskcid1", "diskcid2"},
					VMCreatedAt: time.Date(2016, time.January, 9, 6, 23, 25, 0, time.UTC),

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
					ProcessState: "unresponsive agent",
					Deployment:   "dep",
					ResourcePool: "",
				},
			}
		})

		Context("when listing all deployments", func() {
			Context("when instances are successfully retrieved", func() {
				BeforeEach(func() {
					deployments := []boshdir.Deployment{
						&fakedir.FakeDeployment{
							NameStub: func() string { return "dep1" },
							InstanceInfosStub: func() ([]boshdir.VMInfo, error) {
								infos0 := infos[0]
								infos0.Deployment = "dep1"
								infos0.JobName = "dep1-" + infos0.JobName
								return []boshdir.VMInfo{infos0}, nil
							},
						},
						&fakedir.FakeDeployment{
							NameStub: func() string { return "dep2" },
							InstanceInfosStub: func() ([]boshdir.VMInfo, error) {
								infos0 := infos[0]
								infos0.Deployment = "dep2"
								infos0.JobName = "dep2-" + infos0.JobName
								return []boshdir.VMInfo{infos0}, nil
							},
						},
					}

					director.DeploymentsReturns(deployments, nil)
				})

				It("lists instances for each deployment", func() {
					Expect(act()).ToNot(HaveOccurred())

					Expect(ui.Tables).To(Equal([]boshtbl.Table{
						{
							Title: "Deployment 'dep1'",

							Content: "instances",

							Header: []boshtbl.Header{
								boshtbl.NewHeader("Instance"),
								boshtbl.NewHeader("Process State"),
								boshtbl.NewHeader("AZ"),
								boshtbl.NewHeader("IPs"),
								boshtbl.NewHeader("Deployment"),
							},

							SortBy: []boshtbl.ColumnSort{
								{Column: 0, Asc: true},
								{Column: 1, Asc: true},
							},

							Sections: []boshtbl.Section{
								{
									FirstColumn: boshtbl.NewValueString("dep1-job-name"),
									Rows: [][]boshtbl.Value{
										{
											boshtbl.NewValueString("dep1-job-name"),
											boshtbl.NewValueFmt(boshtbl.NewValueString("in1-process-state"), true),
											boshtbl.ValueString{},
											boshtbl.NewValueStrings([]string{"in1-ip1", "in1-ip2"}),
											boshtbl.NewValueString("dep1"),
										},
									},
								},
							},
						},
						{
							Title: "Deployment 'dep2'",

							Content: "instances",

							Header: []boshtbl.Header{
								boshtbl.NewHeader("Instance"),
								boshtbl.NewHeader("Process State"),
								boshtbl.NewHeader("AZ"),
								boshtbl.NewHeader("IPs"),
								boshtbl.NewHeader("Deployment"),
							},

							SortBy: []boshtbl.ColumnSort{
								{Column: 0, Asc: true},
								{Column: 1, Asc: true},
							},

							Sections: []boshtbl.Section{
								{
									FirstColumn: boshtbl.NewValueString("dep2-job-name"),
									Rows: [][]boshtbl.Value{
										{
											boshtbl.NewValueString("dep2-job-name"),
											boshtbl.NewValueFmt(boshtbl.NewValueString("in1-process-state"), true),
											boshtbl.ValueString{},
											boshtbl.NewValueStrings([]string{"in1-ip1", "in1-ip2"}),
											boshtbl.NewValueString("dep2"),
										},
									},
								},
							},
						},
					}))
				})
			})

			It("returns error if instances cannot be retrieved", func() {
				deployments := []boshdir.Deployment{
					&fakedir.FakeDeployment{
						NameStub:          func() string { return "dep1" },
						InstanceInfosStub: func() ([]boshdir.VMInfo, error) { return nil, errors.New("fake-err") },
					},
				}

				director.DeploymentsReturns(deployments, nil)

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})

			It("returns error if deployments cannot be retrieved", func() {
				director.DeploymentsReturns(nil, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})

		Context("when listing single deployment", func() {
			var (
				deployment *fakedir.FakeDeployment
			)

			BeforeEach(func() {
				instancesOpts.Deployment = "dep"

				deployment = &fakedir.FakeDeployment{
					NameStub:          func() string { return "dep" },
					InstanceInfosStub: func() ([]boshdir.VMInfo, error) { return infos, nil },
				}

				director.FindDeploymentStub = func(name string) (boshdir.Deployment, error) {
					Expect(name).To(Equal("dep"))
					return deployment, nil
				}
			})

			Context("when instances are successfully retrieved", func() {
				It("lists instances for the deployment", func() {
					Expect(act()).ToNot(HaveOccurred())

					Expect(ui.Table).To(Equal(boshtbl.Table{
						Title:   "Deployment 'dep'",
						Content: "instances",

						Header: []boshtbl.Header{
							boshtbl.NewHeader("Instance"),
							boshtbl.NewHeader("Process State"),
							boshtbl.NewHeader("AZ"),
							boshtbl.NewHeader("IPs"),
							boshtbl.NewHeader("Deployment"),
						},

						SortBy: []boshtbl.ColumnSort{
							{Column: 0, Asc: true},
							{Column: 1, Asc: true},
						},

						Sections: []boshtbl.Section{
							{
								FirstColumn: boshtbl.NewValueString("job-name"),
								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueString("job-name"),
										boshtbl.NewValueFmt(boshtbl.NewValueString("in1-process-state"), true),
										boshtbl.ValueString{},
										boshtbl.NewValueStrings([]string{"in1-ip1", "in1-ip2"}),
										boshtbl.NewValueString("dep"),
									},
								},
							},
							{
								FirstColumn: boshtbl.NewValueString("job-name"),
								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueString("job-name"),
										boshtbl.NewValueFmt(boshtbl.NewValueString("in2-process-state"), true),
										boshtbl.NewValueString("in2-az"),
										boshtbl.NewValueStrings([]string{"in2-ip1"}),
										boshtbl.NewValueString("dep"),
									},
								},
							},
							{
								FirstColumn: boshtbl.NewValueString("?"),
								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueString("?"),
										boshtbl.NewValueFmt(boshtbl.NewValueString("unresponsive agent"), true),
										boshtbl.ValueString{},
										boshtbl.ValueStrings{},
										boshtbl.NewValueString("dep"),
									},
								},
							},
						},
					}))
				})

				It("lists instances with processes", func() {
					instancesOpts.Processes = true

					Expect(act()).ToNot(HaveOccurred())

					Expect(ui.Table).To(Equal(boshtbl.Table{
						Title:   "Deployment 'dep'",
						Content: "instances",

						Header: []boshtbl.Header{
							boshtbl.NewHeader("Instance"),
							boshtbl.NewHeader("Process"),
							boshtbl.NewHeader("Process State"),
							boshtbl.NewHeader("AZ"),
							boshtbl.NewHeader("IPs"),
							boshtbl.NewHeader("Deployment"),
						},

						SortBy: []boshtbl.ColumnSort{
							{Column: 0, Asc: true},
							{Column: 1, Asc: true},
						},

						Sections: []boshtbl.Section{
							{
								FirstColumn: boshtbl.NewValueString("job-name"),
								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueString("job-name"),
										boshtbl.ValueString{},
										boshtbl.NewValueFmt(boshtbl.NewValueString("in1-process-state"), true),
										boshtbl.ValueString{},
										boshtbl.NewValueStrings([]string{"in1-ip1", "in1-ip2"}),
										boshtbl.NewValueString("dep"),
									},
									{
										boshtbl.ValueString{},
										boshtbl.NewValueString("in1-proc1-name"),
										boshtbl.NewValueFmt(boshtbl.NewValueString("in1-proc1-state"), true),
										nil,
										nil,
										nil,
									},
									{
										boshtbl.ValueString{},
										boshtbl.NewValueString("in1-proc2-name"),
										boshtbl.NewValueFmt(boshtbl.NewValueString("in1-proc2-state"), true),
										nil,
										nil,
										nil,
									},
								},
							},
							{
								FirstColumn: boshtbl.NewValueString("job-name"),
								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueString("job-name"),
										boshtbl.ValueString{},
										boshtbl.NewValueFmt(boshtbl.NewValueString("in2-process-state"), true),
										boshtbl.NewValueString("in2-az"),
										boshtbl.NewValueStrings([]string{"in2-ip1"}),
										boshtbl.NewValueString("dep"),
									},
									{
										boshtbl.ValueString{},
										boshtbl.NewValueString("in2-proc1-name"),
										boshtbl.NewValueFmt(boshtbl.NewValueString("in2-proc1-state"), true),
										nil,
										nil,
										nil,
									},
								},
							},
							{
								FirstColumn: boshtbl.NewValueString("?"),
								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueString("?"),
										boshtbl.ValueString{},
										boshtbl.NewValueFmt(boshtbl.NewValueString("unresponsive agent"), true),
										boshtbl.ValueString{},
										boshtbl.ValueStrings{},
										boshtbl.NewValueString("dep"),
									},
								},
							},
						},
					}))
				})

				It("lists instances for the deployment including details", func() {
					instancesOpts.Details = true

					Expect(act()).ToNot(HaveOccurred())

					Expect(ui.Table).To(Equal(boshtbl.Table{
						Title:   "Deployment 'dep'",
						Content: "instances",

						Header: []boshtbl.Header{
							boshtbl.NewHeader("Instance"),
							boshtbl.NewHeader("Process State"),
							boshtbl.NewHeader("AZ"),
							boshtbl.NewHeader("IPs"),
							boshtbl.NewHeader("Deployment"),
							boshtbl.NewHeader("State"),
							boshtbl.NewHeader("VM CID"),
							boshtbl.NewHeader("VM Type"),
							boshtbl.NewHeader("Disk CIDs"),
							boshtbl.NewHeader("Agent ID"),
							boshtbl.NewHeader("Index"),
							boshtbl.NewHeader("Bootstrap"),
							boshtbl.NewHeader("Ignore"),
						},

						SortBy: []boshtbl.ColumnSort{
							{Column: 0, Asc: true},
							{Column: 1, Asc: true},
						},

						Sections: []boshtbl.Section{
							{
								FirstColumn: boshtbl.NewValueString("job-name"),
								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueString("job-name"),
										boshtbl.NewValueFmt(boshtbl.NewValueString("in1-process-state"), true),
										boshtbl.ValueString{},
										boshtbl.NewValueStrings([]string{"in1-ip1", "in1-ip2"}),
										boshtbl.NewValueString("dep"),
										boshtbl.NewValueString("in1-state"),
										boshtbl.NewValueString("in1-cid"),
										boshtbl.NewValueString("in1-rp"),
										boshtbl.NewValueStrings([]string{"diskcid1", "diskcid2"}),
										boshtbl.NewValueString("in1-agent-id"),
										boshtbl.NewValueInt(1),
										boshtbl.NewValueBool(false),
										boshtbl.NewValueBool(true),
									},
								},
							},
							{
								FirstColumn: boshtbl.NewValueString("job-name"),
								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueString("job-name"),
										boshtbl.NewValueFmt(boshtbl.NewValueString("in2-process-state"), true),
										boshtbl.NewValueString("in2-az"),
										boshtbl.NewValueStrings([]string{"in2-ip1"}),
										boshtbl.NewValueString("dep"),
										boshtbl.NewValueString("in2-state"),
										boshtbl.NewValueString("in2-cid"),
										boshtbl.NewValueString("in2-rp"),
										boshtbl.NewValueStrings([]string{"diskcid1", "diskcid2"}),
										boshtbl.NewValueString("in2-agent-id"),
										boshtbl.NewValueInt(2),
										boshtbl.NewValueBool(false),
										boshtbl.NewValueBool(false),
									},
								},
							},
							{
								FirstColumn: boshtbl.NewValueString("?"),
								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueString("?"),
										boshtbl.NewValueFmt(boshtbl.NewValueString("unresponsive agent"), true),
										boshtbl.ValueString{},
										boshtbl.ValueStrings{},
										boshtbl.NewValueString("dep"),
										boshtbl.NewValueString(""),
										boshtbl.ValueString{},
										boshtbl.ValueString{},
										boshtbl.ValueStrings{},
										boshtbl.ValueString{},
										boshtbl.NewValueInt(0),
										boshtbl.NewValueBool(false),
										boshtbl.NewValueBool(false),
									},
								},
							},
						},
					}))
				})

				It("lists instances for the deployment including vitals and processes", func() {
					instancesOpts.Vitals = true
					instancesOpts.Processes = true

					Expect(act()).ToNot(HaveOccurred())

					Expect(ui.Table).To(Equal(boshtbl.Table{
						Title:   "Deployment 'dep'",
						Content: "instances",

						Header: []boshtbl.Header{
							boshtbl.NewHeader("Instance"),
							boshtbl.NewHeader("Process"),
							boshtbl.NewHeader("Process State"),
							boshtbl.NewHeader("AZ"),
							boshtbl.NewHeader("IPs"),
							boshtbl.NewHeader("Deployment"),
							boshtbl.NewHeader("VM Created At"),
							boshtbl.NewHeader("Uptime"),
							boshtbl.NewHeader("Load\n(1m, 5m, 15m)"),
							boshtbl.NewHeader("CPU\nTotal"),
							boshtbl.NewHeader("CPU\nUser"),
							boshtbl.NewHeader("CPU\nSys"),
							boshtbl.NewHeader("CPU\nWait"),
							boshtbl.NewHeader("Memory\nUsage"),
							boshtbl.NewHeader("Swap\nUsage"),
							boshtbl.NewHeader("System\nDisk Usage"),
							boshtbl.NewHeader("Ephemeral\nDisk Usage"),
							boshtbl.NewHeader("Persistent\nDisk Usage"),
						},

						SortBy: []boshtbl.ColumnSort{
							{Column: 0, Asc: true},
							{Column: 1, Asc: true},
						},

						Sections: []boshtbl.Section{
							{
								FirstColumn: boshtbl.NewValueString("job-name"),
								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueString("job-name"),
										boshtbl.ValueString{},
										boshtbl.NewValueFmt(boshtbl.NewValueString("in1-process-state"), true),
										boshtbl.ValueString{},
										boshtbl.NewValueStrings([]string{"in1-ip1", "in1-ip2"}),
										boshtbl.NewValueString("dep"),
										boshtbl.NewValueTime(time.Date(2016, time.January, 9, 6, 23, 25, 0, time.UTC)),
										cmd.ValueUptime{},
										boshtbl.NewValueString("0.02, 0.06, 0.11"),
										cmd.ValueCPUTotal{},
										cmd.NewValueStringPercent("1.2"),
										cmd.NewValueStringPercent("0.3"),
										cmd.NewValueStringPercent("2.1"),
										cmd.ValueMemSize{Size: boshdir.VMInfoVitalsMemSize{Percent: "20", KB: "2000"}},
										cmd.ValueMemSize{Size: boshdir.VMInfoVitalsMemSize{Percent: "21", KB: "2100"}},
										cmd.ValueDiskSize{Size: boshdir.VMInfoVitalsDiskSize{Percent: "35"}},
										cmd.ValueDiskSize{Size: boshdir.VMInfoVitalsDiskSize{Percent: "45"}},
										cmd.ValueDiskSize{Size: boshdir.VMInfoVitalsDiskSize{Percent: "55"}},
									},
									{
										boshtbl.ValueString{},
										boshtbl.NewValueString("in1-proc1-name"),
										boshtbl.NewValueFmt(boshtbl.NewValueString("in1-proc1-state"), true),
										nil,
										nil,
										nil,
										nil,
										cmd.ValueUptime{Secs: &procUptime},
										nil,
										cmd.ValueCPUTotal{Total: &procCPUTotal},
										nil,
										nil,
										nil,
										cmd.ValueMemIntSize{Size: boshdir.VMInfoVitalsMemIntSize{Percent: &procMemPercent, KB: &procMemKB}},
										nil,
										nil,
										nil,
										nil,
									},
									{
										boshtbl.ValueString{},
										boshtbl.NewValueString("in1-proc2-name"),
										boshtbl.NewValueFmt(boshtbl.NewValueString("in1-proc2-state"), true),
										nil,
										nil,
										nil,
										nil,
										cmd.ValueUptime{},
										nil,
										cmd.ValueCPUTotal{},
										nil,
										nil,
										nil,
										cmd.ValueMemIntSize{},
										nil,
										nil,
										nil,
										nil,
									},
								},
							},
							{
								FirstColumn: boshtbl.NewValueString("job-name"),
								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueString("job-name"),
										boshtbl.ValueString{},
										boshtbl.NewValueFmt(boshtbl.NewValueString("in2-process-state"), true),
										boshtbl.NewValueString("in2-az"),
										boshtbl.NewValueStrings([]string{"in2-ip1"}),
										boshtbl.NewValueString("dep"),
										boshtbl.NewValueTime(time.Date(2016, time.January, 9, 6, 23, 25, 0, time.UTC)),
										cmd.ValueUptime{},
										boshtbl.NewValueString("0.52, 0.56, 0.51"),
										cmd.ValueCPUTotal{},
										cmd.NewValueStringPercent("51.2"),
										cmd.NewValueStringPercent("50.3"),
										cmd.NewValueStringPercent("52.1"),
										cmd.ValueMemSize{Size: boshdir.VMInfoVitalsMemSize{Percent: "60", KB: "6000"}},
										cmd.ValueMemSize{Size: boshdir.VMInfoVitalsMemSize{Percent: "61", KB: "6100"}},
										cmd.ValueDiskSize{Size: boshdir.VMInfoVitalsDiskSize{Percent: "75"}},
										cmd.ValueDiskSize{Size: boshdir.VMInfoVitalsDiskSize{Percent: "85"}},
										cmd.ValueDiskSize{Size: boshdir.VMInfoVitalsDiskSize{Percent: "95"}},
									},
									{
										boshtbl.ValueString{},
										boshtbl.NewValueString("in2-proc1-name"),
										boshtbl.NewValueFmt(boshtbl.NewValueString("in2-proc1-state"), true),
										nil,
										nil,
										nil,
										nil,
										cmd.ValueUptime{},
										nil,
										cmd.ValueCPUTotal{},
										nil,
										nil,
										nil,
										cmd.ValueMemIntSize{},
										nil,
										nil,
										nil,
										nil,
									},
								},
							},
							{
								FirstColumn: boshtbl.NewValueString("?"),
								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueString("?"),
										boshtbl.ValueString{},
										boshtbl.NewValueFmt(boshtbl.NewValueString("unresponsive agent"), true),
										boshtbl.ValueString{},
										boshtbl.ValueStrings{},
										boshtbl.NewValueString("dep"),
										boshtbl.NewValueTime(time.Time{}.UTC()),
										cmd.ValueUptime{},
										boshtbl.ValueString{},
										cmd.ValueCPUTotal{},
										cmd.NewValueStringPercent(""),
										cmd.NewValueStringPercent(""),
										cmd.NewValueStringPercent(""),
										cmd.ValueMemSize{},
										cmd.ValueMemSize{},
										cmd.ValueDiskSize{},
										cmd.ValueDiskSize{},
										cmd.ValueDiskSize{},
									},
								},
							},
						},
					}))
				})

				It("lists failing (non-running) instances", func() {
					instancesOpts.Failing = true

					// Hides second VM
					infos[1].ProcessState = "running"
					infos[1].Processes[0].State = "running"

					Expect(act()).ToNot(HaveOccurred())

					Expect(ui.Table).To(Equal(boshtbl.Table{
						Title:   "Deployment 'dep'",
						Content: "instances",

						Header: []boshtbl.Header{
							boshtbl.NewHeader("Instance"),
							boshtbl.NewHeader("Process State"),
							boshtbl.NewHeader("AZ"),
							boshtbl.NewHeader("IPs"),
							boshtbl.NewHeader("Deployment"),
						},

						SortBy: []boshtbl.ColumnSort{
							{Column: 0, Asc: true},
							{Column: 1, Asc: true},
						},

						Sections: []boshtbl.Section{
							{
								FirstColumn: boshtbl.NewValueString("job-name"),
								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueString("job-name"),
										boshtbl.NewValueFmt(boshtbl.NewValueString("in1-process-state"), true),
										boshtbl.ValueString{},
										boshtbl.NewValueStrings([]string{"in1-ip1", "in1-ip2"}),
										boshtbl.NewValueString("dep"),
									},
								},
							},
							{
								FirstColumn: boshtbl.NewValueString("?"),
								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueString("?"),
										boshtbl.NewValueFmt(boshtbl.NewValueString("unresponsive agent"), true),
										boshtbl.ValueString{},
										boshtbl.ValueStrings{},
										boshtbl.NewValueString("dep"),
									},
								},
							},
						},
					}))
				})

				It("includes failing processes when listing failing (non-running) instances and processes", func() {
					instancesOpts.Failing = true
					instancesOpts.Processes = true

					// Hides first process in the first VM
					infos[0].Processes[0].State = "running"

					// Hides second VM completely
					infos[1].ProcessState = "running"
					infos[1].Processes[0].State = "running"

					Expect(act()).ToNot(HaveOccurred())

					Expect(ui.Table).To(Equal(boshtbl.Table{
						Title:   "Deployment 'dep'",
						Content: "instances",

						Header: []boshtbl.Header{
							boshtbl.NewHeader("Instance"),
							boshtbl.NewHeader("Process"),
							boshtbl.NewHeader("Process State"),
							boshtbl.NewHeader("AZ"),
							boshtbl.NewHeader("IPs"),
							boshtbl.NewHeader("Deployment"),
						},

						SortBy: []boshtbl.ColumnSort{
							{Column: 0, Asc: true},
							{Column: 1, Asc: true},
						},

						Sections: []boshtbl.Section{
							{
								FirstColumn: boshtbl.NewValueString("job-name"),
								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueString("job-name"),
										boshtbl.ValueString{},
										boshtbl.NewValueFmt(boshtbl.NewValueString("in1-process-state"), true),
										boshtbl.ValueString{},
										boshtbl.NewValueStrings([]string{"in1-ip1", "in1-ip2"}),
										boshtbl.NewValueString("dep"),
									},
									{
										boshtbl.ValueString{},
										boshtbl.NewValueString("in1-proc2-name"),
										boshtbl.NewValueFmt(boshtbl.NewValueString("in1-proc2-state"), true),
										nil,
										nil,
										nil,
									},
								},
							},
							{
								FirstColumn: boshtbl.NewValueString("?"),
								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueString("?"),
										boshtbl.ValueString{},
										boshtbl.NewValueFmt(boshtbl.NewValueString("unresponsive agent"), true),
										boshtbl.ValueString{},
										boshtbl.ValueStrings{},
										boshtbl.NewValueString("dep"),
									},
								},
							},
						},
					}))
				})
			})

			It("returns error if instances cannot be retrieved", func() {
				deployment.InstanceInfosReturns(nil, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})

			It("returns error if finding deployment fails", func() {
				director.FindDeploymentReturns(nil, errors.New("fake-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})

		Context("when listing multiple deployments", func() {
			BeforeEach(func() {
				command = cmd.NewInstancesCmd(ui, director, 5)
			})

			It("retrieves deployment vms in parallel", func() {
				dep1 := &fakedir.FakeDeployment{
					NameStub: func() string { return "dep1" },
					InstanceInfosStub: func() ([]boshdir.VMInfo, error) {
						time.Sleep(1500 * time.Millisecond)
						return infos, nil
					},
				}
				dep2 := &fakedir.FakeDeployment{
					NameStub: func() string { return "dep2" },
					InstanceInfosStub: func() ([]boshdir.VMInfo, error) {
						time.Sleep(1500 * time.Millisecond)
						return infos, nil
					},
				}
				deployments := []boshdir.Deployment{
					dep1,
					dep2,
				}

				director.DeploymentsReturns(deployments, nil)
				startTime := time.Now()
				err := act()
				cmdDuration := time.Since(startTime)
				Expect(err).To(BeNil())
				Expect(int64(cmdDuration / time.Millisecond)).To(BeNumerically("<", 2000))
				Expect(dep1.InstanceInfosCallCount()).To(Equal(1))
				Expect(dep2.InstanceInfosCallCount()).To(Equal(1))
			})
		})

		Context("when fetching vms infos from subset of deployment fail", func() {
			It("returns instance info and errors", func() {
				vmError := bosherr.Error("failed")
				dep1 := &fakedir.FakeDeployment{
					NameStub: func() string { return "dep1" },
					InstanceInfosStub: func() ([]boshdir.VMInfo, error) {
						infos[0].Deployment = "dep1"
						infos[1].Deployment = "dep1"
						infos[2].Deployment = "dep1"
						return infos, nil
					},
				}
				dep2 := &fakedir.FakeDeployment{
					NameStub: func() string { return "dep2" },
					InstanceInfosStub: func() ([]boshdir.VMInfo, error) {
						return nil, vmError
					},
				}
				deployments := []boshdir.Deployment{
					dep1,
					dep2,
				}
				director.DeploymentsReturns(deployments, nil)
				err := act()
				Expect(err).To(Equal(bosherr.NewMultiError(vmError)))
				Expect(ui.Table).To(Equal(boshtbl.Table{
					Title: "Deployment 'dep1'",

					Content: "instances",

					Header: []boshtbl.Header{
						boshtbl.NewHeader("Instance"),
						boshtbl.NewHeader("Process State"),
						boshtbl.NewHeader("AZ"),
						boshtbl.NewHeader("IPs"),
						boshtbl.NewHeader("Deployment"),
					},

					SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}, {Column: 1, Asc: true}},

					Sections: []boshtbl.Section{
						{
							FirstColumn: boshtbl.NewValueString("job-name"),
							Rows: [][]boshtbl.Value{
								{
									boshtbl.NewValueString("job-name"),
									boshtbl.NewValueFmt(boshtbl.NewValueString("in1-process-state"), true),
									boshtbl.ValueString{},
									boshtbl.NewValueStrings([]string{"in1-ip1", "in1-ip2"}),
									boshtbl.NewValueString("dep1"),
								},
							},
						}, {
							FirstColumn: boshtbl.NewValueString("job-name"),
							Rows: [][]boshtbl.Value{
								{
									boshtbl.NewValueString("job-name"),
									boshtbl.NewValueFmt(boshtbl.NewValueString("in2-process-state"), true),
									boshtbl.NewValueString("in2-az"),
									boshtbl.NewValueStrings([]string{"in2-ip1"}),
									boshtbl.NewValueString("dep1"),
								},
							},
						}, {
							FirstColumn: boshtbl.NewValueString("?"),
							Rows: [][]boshtbl.Value{
								{
									boshtbl.NewValueString("?"),
									boshtbl.NewValueFmt(boshtbl.NewValueString("unresponsive agent"), true),
									boshtbl.ValueString{},
									boshtbl.ValueStrings{},
									boshtbl.NewValueString("dep1"),
								},
							},
						},
					},
				}))
				Expect(dep1.InstanceInfosCallCount()).To(Equal(1))
				Expect(dep2.InstanceInfosCallCount()).To(Equal(1))
			})
		})
	})
})
