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

var _ = Describe("VMsCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  cmd.VMsCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewVMsCmd(ui, director, 1)
	})

	Describe("Run", func() {
		var (
			vMsOpts opts.VMsOpts
			infos   []boshdir.VMInfo
		)

		BeforeEach(func() {
			vMsOpts = opts.VMsOpts{}
		})

		act := func() error { return command.Run(vMsOpts) }

		BeforeEach(func() {
			index1 := 1
			index2 := 2

			var cloudProperties interface{} = map[string]string{"instance_type": "m1.small"}

			t := true
			f := false

			infos = []boshdir.VMInfo{
				{
					JobName:      "job-name",
					Index:        &index1,
					ProcessState: "in1-process-state",
					ResourcePool: "in1-rp",
					Active:       &t,

					IPs:        []string{"in1-ip1", "in1-ip2"},
					Deployment: "dep",

					VMID:               "in1-cid",
					AgentID:            "in1-agent-id",
					ResurrectionPaused: false,
					Ignore:             false,
					DiskIDs:            []string{"diskcid1", "diskcid2"},
					VMCreatedAt:        time.Date(2016, time.January, 9, 6, 23, 25, 0, time.UTC),
					CloudProperties:    cloudProperties,
					Stemcell:           boshdir.VmInfoStemcell{Name: "stemcell", Version: "version", ApiVersion: 1},

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
				},
				{
					JobName:      "job-name",
					Index:        &index2,
					ProcessState: "in2-process-state",
					AZ:           "in2-az",
					ResourcePool: "in2-rp",
					Active:       &f,

					IPs:        []string{"in2-ip1"},
					Deployment: "dep",

					VMID:               "in2-cid",
					AgentID:            "in2-agent-id",
					ResurrectionPaused: true,
					Ignore:             true,
					DiskIDs:            []string{"diskcid1", "diskcid2"},
					VMCreatedAt:        time.Date(2016, time.January, 9, 6, 23, 25, 0, time.UTC),
					CloudProperties:    cloudProperties,
					Stemcell:           boshdir.VmInfoStemcell{Name: "stemcell", Version: "version", ApiVersion: 1},

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
				},
				{
					JobName:      "",
					Index:        nil,
					ProcessState: "unresponsive agent",
					ResourcePool: "",
					Deployment:   "dep",
				},
			}
		})

		Context("when listing all deployments", func() {
			Context("when VMs are successfully retrieved", func() {
				BeforeEach(func() {
					deployments := []boshdir.Deployment{
						&fakedir.FakeDeployment{
							NameStub:    func() string { return "dep1" },
							VMInfosStub: func() ([]boshdir.VMInfo, error) { return infos, nil },
						},
					}

					director.DeploymentsReturns(deployments, nil)
				})

				It("lists VMs for the deployment", func() {
					Expect(act()).ToNot(HaveOccurred())

					Expect(ui.Table).To(Equal(boshtbl.Table{
						Title: "Deployment 'dep1'",

						Content: "vms",

						Header: []boshtbl.Header{
							boshtbl.NewHeader("Instance"),
							boshtbl.NewHeader("Process State"),
							boshtbl.NewHeader("AZ"),
							boshtbl.NewHeader("IPs"),
							boshtbl.NewHeader("VM CID"),
							boshtbl.NewHeader("VM Type"),
							boshtbl.NewHeader("Active"),
							boshtbl.NewHeader("Stemcell"),
						},

						SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

						Rows: [][]boshtbl.Value{
							{
								boshtbl.NewValueString("job-name"),
								boshtbl.NewValueFmt(boshtbl.NewValueString("in1-process-state"), true),
								boshtbl.ValueString{},
								boshtbl.NewValueStrings([]string{"in1-ip1", "in1-ip2"}),
								boshtbl.NewValueString("in1-cid"),
								boshtbl.NewValueString("in1-rp"),
								boshtbl.NewValueString("true"),
								boshtbl.NewValueString("stemcell/version"),
							},
							{
								boshtbl.NewValueString("job-name"),
								boshtbl.NewValueFmt(boshtbl.NewValueString("in2-process-state"), true),
								boshtbl.NewValueString("in2-az"),
								boshtbl.NewValueStrings([]string{"in2-ip1"}),
								boshtbl.NewValueString("in2-cid"),
								boshtbl.NewValueString("in2-rp"),
								boshtbl.NewValueString("false"),
								boshtbl.NewValueString("stemcell/version"),
							},
							{
								boshtbl.NewValueString("?"),
								boshtbl.NewValueFmt(boshtbl.NewValueString("unresponsive agent"), true),
								boshtbl.ValueString{},
								boshtbl.ValueStrings{},
								boshtbl.ValueString{},
								boshtbl.ValueString{},
								boshtbl.ValueString{S: "-"},
								boshtbl.ValueString{S: "-"},
							},
						},
					}))
				})

				It("lists VMs for the deployment including vitals", func() {
					vMsOpts.Vitals = true

					Expect(act()).ToNot(HaveOccurred())

					Expect(ui.Table).To(Equal(boshtbl.Table{
						Title: "Deployment 'dep1'",

						Content: "vms",

						Header: []boshtbl.Header{
							boshtbl.NewHeader("Instance"),
							boshtbl.NewHeader("Process State"),
							boshtbl.NewHeader("AZ"),
							boshtbl.NewHeader("IPs"),
							boshtbl.NewHeader("VM CID"),
							boshtbl.NewHeader("VM Type"),
							boshtbl.NewHeader("Active"),
							boshtbl.NewHeader("Stemcell"),
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

						SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

						Rows: [][]boshtbl.Value{
							{
								boshtbl.NewValueString("job-name"),
								boshtbl.NewValueFmt(boshtbl.NewValueString("in1-process-state"), true),
								boshtbl.ValueString{},
								boshtbl.NewValueStrings([]string{"in1-ip1", "in1-ip2"}),
								boshtbl.NewValueString("in1-cid"),
								boshtbl.NewValueString("in1-rp"),
								boshtbl.NewValueString("true"),
								boshtbl.NewValueString("stemcell/version"),
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
								boshtbl.NewValueString("job-name"),
								boshtbl.NewValueFmt(boshtbl.NewValueString("in2-process-state"), true),
								boshtbl.NewValueString("in2-az"),
								boshtbl.NewValueStrings([]string{"in2-ip1"}),
								boshtbl.NewValueString("in2-cid"),
								boshtbl.NewValueString("in2-rp"),
								boshtbl.NewValueString("false"),
								boshtbl.NewValueString("stemcell/version"),
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
								boshtbl.NewValueString("?"),
								boshtbl.NewValueFmt(boshtbl.NewValueString("unresponsive agent"), true),
								boshtbl.ValueString{},
								boshtbl.ValueStrings{},
								boshtbl.ValueString{},
								boshtbl.ValueString{},
								boshtbl.ValueString{S: "-"},
								boshtbl.ValueString{S: "-"},
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
					}))
				})

				It("lists VMs for the deployment including cloud properties", func() {
					vMsOpts.CloudProperties = true

					Expect(act()).ToNot(HaveOccurred())

					Expect(ui.Table).To(Equal(boshtbl.Table{
						Title: "Deployment 'dep1'",

						Content: "vms",

						Header: []boshtbl.Header{
							boshtbl.NewHeader("Instance"),
							boshtbl.NewHeader("Process State"),
							boshtbl.NewHeader("AZ"),
							boshtbl.NewHeader("IPs"),
							boshtbl.NewHeader("VM CID"),
							boshtbl.NewHeader("VM Type"),
							boshtbl.NewHeader("Active"),
							boshtbl.NewHeader("Stemcell"),
							boshtbl.NewHeader("Cloud Properties"),
						},

						SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

						Rows: [][]boshtbl.Value{
							{
								boshtbl.NewValueString("job-name"),
								boshtbl.NewValueFmt(boshtbl.NewValueString("in1-process-state"), true),
								boshtbl.ValueString{},
								boshtbl.NewValueStrings([]string{"in1-ip1", "in1-ip2"}),
								boshtbl.NewValueString("in1-cid"),
								boshtbl.NewValueString("in1-rp"),
								boshtbl.NewValueString("true"),
								boshtbl.NewValueString("stemcell/version"),
								boshtbl.NewValueInterface(map[string]string{"instance_type": "m1.small"}),
							},
							{
								boshtbl.NewValueString("job-name"),
								boshtbl.NewValueFmt(boshtbl.NewValueString("in2-process-state"), true),
								boshtbl.NewValueString("in2-az"),
								boshtbl.NewValueStrings([]string{"in2-ip1"}),
								boshtbl.NewValueString("in2-cid"),
								boshtbl.NewValueString("in2-rp"),
								boshtbl.NewValueString("false"),
								boshtbl.NewValueString("stemcell/version"),
								boshtbl.NewValueInterface(map[string]string{"instance_type": "m1.small"}),
							},
							{
								boshtbl.NewValueString("?"),
								boshtbl.NewValueFmt(boshtbl.NewValueString("unresponsive agent"), true),
								boshtbl.ValueString{},
								boshtbl.ValueStrings{},
								boshtbl.ValueString{},
								boshtbl.ValueString{},
								boshtbl.ValueString{S: "-"},
								boshtbl.ValueString{S: "-"},
								boshtbl.ValueInterface{},
							},
						},
					}))
				})
			})

			Context("when legacy", func() {
				BeforeEach(func() {
					infos[0].Active = nil
					infos[1].Active = nil
					deployments := []boshdir.Deployment{
						&fakedir.FakeDeployment{
							NameStub:    func() string { return "dep1" },
							VMInfosStub: func() ([]boshdir.VMInfo, error) { return infos, nil },
						},
					}

					director.DeploymentsReturns(deployments, nil)
				})

				It("lists VMs for the deployment with Active status '-' for a legacy director", func() {
					Expect(act()).ToNot(HaveOccurred())

					Expect(ui.Table).To(Equal(boshtbl.Table{
						Title: "Deployment 'dep1'",

						Content: "vms",

						Header: []boshtbl.Header{
							boshtbl.NewHeader("Instance"),
							boshtbl.NewHeader("Process State"),
							boshtbl.NewHeader("AZ"),
							boshtbl.NewHeader("IPs"),
							boshtbl.NewHeader("VM CID"),
							boshtbl.NewHeader("VM Type"),
							boshtbl.NewHeader("Active"),
							boshtbl.NewHeader("Stemcell"),
						},

						SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

						Rows: [][]boshtbl.Value{
							{
								boshtbl.NewValueString("job-name"),
								boshtbl.NewValueFmt(boshtbl.NewValueString("in1-process-state"), true),
								boshtbl.ValueString{},
								boshtbl.NewValueStrings([]string{"in1-ip1", "in1-ip2"}),
								boshtbl.NewValueString("in1-cid"),
								boshtbl.NewValueString("in1-rp"),
								boshtbl.NewValueString("-"),
								boshtbl.NewValueString("stemcell/version"),
							},
							{
								boshtbl.NewValueString("job-name"),
								boshtbl.NewValueFmt(boshtbl.NewValueString("in2-process-state"), true),
								boshtbl.NewValueString("in2-az"),
								boshtbl.NewValueStrings([]string{"in2-ip1"}),
								boshtbl.NewValueString("in2-cid"),
								boshtbl.NewValueString("in2-rp"),
								boshtbl.NewValueString("-"),
								boshtbl.NewValueString("stemcell/version"),
							},
							{
								boshtbl.NewValueString("?"),
								boshtbl.NewValueFmt(boshtbl.NewValueString("unresponsive agent"), true),
								boshtbl.ValueString{},
								boshtbl.ValueStrings{},
								boshtbl.ValueString{},
								boshtbl.ValueString{},
								boshtbl.ValueString{S: "-"},
								boshtbl.ValueString{S: "-"},
							},
						},
					}))
				})
			})

			It("returns error if VMs cannot be retrieved", func() {
				deployments := []boshdir.Deployment{
					&fakedir.FakeDeployment{
						NameStub:    func() string { return "dep1" },
						VMInfosStub: func() ([]boshdir.VMInfo, error) { return nil, errors.New("fake-err") },
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
			BeforeEach(func() {
				vMsOpts.Deployment = "dep1"
			})

			It("lists VMs for the deployment", func() {
				deployment := &fakedir.FakeDeployment{
					NameStub:    func() string { return "dep1" },
					VMInfosStub: func() ([]boshdir.VMInfo, error) { return infos, nil },
				}

				director.FindDeploymentReturns(deployment, nil)

				Expect(act()).ToNot(HaveOccurred())

				Expect(ui.Table).To(Equal(boshtbl.Table{
					Title: "Deployment 'dep1'",

					Content: "vms",

					Header: []boshtbl.Header{
						boshtbl.NewHeader("Instance"),
						boshtbl.NewHeader("Process State"),
						boshtbl.NewHeader("AZ"),
						boshtbl.NewHeader("IPs"),
						boshtbl.NewHeader("VM CID"),
						boshtbl.NewHeader("VM Type"),
						boshtbl.NewHeader("Active"),
						boshtbl.NewHeader("Stemcell"),
					},

					SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

					Rows: [][]boshtbl.Value{
						{
							boshtbl.NewValueString("job-name"),
							boshtbl.NewValueFmt(boshtbl.NewValueString("in1-process-state"), true),
							boshtbl.ValueString{},
							boshtbl.NewValueStrings([]string{"in1-ip1", "in1-ip2"}),
							boshtbl.NewValueString("in1-cid"),
							boshtbl.NewValueString("in1-rp"),
							boshtbl.NewValueString("true"),
							boshtbl.NewValueString("stemcell/version"),
						},
						{
							boshtbl.NewValueString("job-name"),
							boshtbl.NewValueFmt(boshtbl.NewValueString("in2-process-state"), true),
							boshtbl.NewValueString("in2-az"),
							boshtbl.NewValueStrings([]string{"in2-ip1"}),
							boshtbl.NewValueString("in2-cid"),
							boshtbl.NewValueString("in2-rp"),
							boshtbl.NewValueString("false"),
							boshtbl.NewValueString("stemcell/version"),
						},
						{
							boshtbl.NewValueString("?"),
							boshtbl.NewValueFmt(boshtbl.NewValueString("unresponsive agent"), true),
							boshtbl.ValueString{},
							boshtbl.ValueStrings{},
							boshtbl.ValueString{},
							boshtbl.ValueString{},
							boshtbl.ValueString{S: "-"},
							boshtbl.ValueString{S: "-"},
						},
					},
				}))
			})

			It("returns error if VMs cannot be retrieved", func() {
				deployment := &fakedir.FakeDeployment{
					NameStub:    func() string { return "dep1" },
					VMInfosStub: func() ([]boshdir.VMInfo, error) { return nil, errors.New("fake-err") },
				}

				director.FindDeploymentReturns(deployment, nil)

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
				command = cmd.NewVMsCmd(ui, director, 5)
			})

			It("retrieves deployment vms in parallel", func() {
				dep1 := &fakedir.FakeDeployment{
					NameStub: func() string { return "dep1" },
					VMInfosStub: func() ([]boshdir.VMInfo, error) {
						time.Sleep(1500 * time.Millisecond)

						return infos, nil
					},
				}
				dep2 := &fakedir.FakeDeployment{
					NameStub: func() string { return "dep2" },
					VMInfosStub: func() ([]boshdir.VMInfo, error) {
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
				Expect(dep1.VMInfosCallCount()).To(Equal(1))
				Expect(dep2.VMInfosCallCount()).To(Equal(1))
			})

			Context("when fetching vms infos from subset of deployment fail", func() {
				It("returns vm info and errors", func() {
					vmError := bosherr.Error("failed")
					dep1 := &fakedir.FakeDeployment{
						NameStub: func() string { return "dep1" },
						VMInfosStub: func() ([]boshdir.VMInfo, error) {
							return infos, nil
						},
					}
					dep2 := &fakedir.FakeDeployment{
						NameStub: func() string { return "dep2" },
						VMInfosStub: func() ([]boshdir.VMInfo, error) {
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

						Content: "vms",

						Header: []boshtbl.Header{
							boshtbl.NewHeader("Instance"),
							boshtbl.NewHeader("Process State"),
							boshtbl.NewHeader("AZ"),
							boshtbl.NewHeader("IPs"),
							boshtbl.NewHeader("VM CID"),
							boshtbl.NewHeader("VM Type"),
							boshtbl.NewHeader("Active"),
							boshtbl.NewHeader("Stemcell"),
						},

						SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

						Rows: [][]boshtbl.Value{
							{
								boshtbl.NewValueString("job-name"),
								boshtbl.NewValueFmt(boshtbl.NewValueString("in1-process-state"), true),
								boshtbl.ValueString{},
								boshtbl.NewValueStrings([]string{"in1-ip1", "in1-ip2"}),
								boshtbl.NewValueString("in1-cid"),
								boshtbl.NewValueString("in1-rp"),
								boshtbl.NewValueString("true"),
								boshtbl.NewValueString("stemcell/version"),
							},
							{
								boshtbl.NewValueString("job-name"),
								boshtbl.NewValueFmt(boshtbl.NewValueString("in2-process-state"), true),
								boshtbl.NewValueString("in2-az"),
								boshtbl.NewValueStrings([]string{"in2-ip1"}),
								boshtbl.NewValueString("in2-cid"),
								boshtbl.NewValueString("in2-rp"),
								boshtbl.NewValueString("false"),
								boshtbl.NewValueString("stemcell/version"),
							},
							{
								boshtbl.NewValueString("?"),
								boshtbl.NewValueFmt(boshtbl.NewValueString("unresponsive agent"), true),
								boshtbl.ValueString{},
								boshtbl.ValueStrings{},
								boshtbl.ValueString{},
								boshtbl.ValueString{},
								boshtbl.ValueString{S: "-"},
								boshtbl.ValueString{S: "-"},
							},
						},
					}))
					Expect(dep1.VMInfosCallCount()).To(Equal(1))
					Expect(dep2.VMInfosCallCount()).To(Equal(1))
				})
			})
		})
	})
})
