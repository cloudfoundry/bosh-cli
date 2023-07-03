package cmd_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd"
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

func createResolution(name, plan string) boshdir.ProblemResolution {
	return boshdir.ProblemResolution{
		Name: &name,
		Plan: plan,
	}
}

var _ = Describe("CreateRecoveryPlanCmd", func() {
	skipResolution := createResolution("ignore", "Skip for now")
	recreateResolution := createResolution("recreate_vm", "Recreate VM")
	rebootResolution := createResolution("reboot_vm", "Reboot VM")
	deleteVmReferenceResolution := createResolution("delete_vm_reference", "Delete VM reference")
	deleteDiskReferenceResolution := createResolution("delete_disk_reference", "Delete disk reference (DANGEROUS!)")
	reattachDiskResolution := createResolution("reattach_disk", "Reattach disk to instance")
	reattachDiskAndRebootResolution := createResolution("reattach_disk_and_reboot", "Reattach disk and reboot instance")

	var (
		deployment *fakedir.FakeDeployment
		ui         *fakeui.FakeUI
		fakeFS     *fakesys.FakeFileSystem
		command    CreateRecoveryPlanCmd
	)

	BeforeEach(func() {
		deployment = &fakedir.FakeDeployment{}
		ui = &fakeui.FakeUI{}
		fakeFS = fakesys.NewFakeFileSystem()
		command = NewCreateRecoveryPlanCmd(deployment, ui, fakeFS)
	})

	Describe("Run", func() {
		var (
			opts         CreateRecoveryPlanOpts
			severalProbs []boshdir.Problem
		)

		BeforeEach(func() {
			opts = CreateRecoveryPlanOpts{
				Args: CreateRecoveryPlanArgs{
					RecoveryPlan: FileArg{
						ExpandedPath: "/tmp/foo.yml",
						FS:           fakeFS,
					},
				},
			}

			severalProbs = []boshdir.Problem{
				{
					ID: 3,

					Type:          "unresponsive_agent",
					Description:   "problem1-desc",
					InstanceGroup: "diego_cell",

					Resolutions: []boshdir.ProblemResolution{
						skipResolution,
						recreateResolution,
						deleteVmReferenceResolution,
					},
				},
				{
					ID: 4,

					Type:          "missing_vm",
					Description:   "problem2-desc",
					InstanceGroup: "router",

					Resolutions: []boshdir.ProblemResolution{
						skipResolution,
						recreateResolution,
						rebootResolution,
						deleteDiskReferenceResolution,
					},
				},
				{
					ID: 5,

					Type:          "mount_info_mismatch",
					Description:   "problem3-desc",
					InstanceGroup: "router",

					Resolutions: []boshdir.ProblemResolution{
						skipResolution,
						reattachDiskResolution,
						reattachDiskAndRebootResolution,
					},
				},
			}
		})

		act := func() error { return command.Run(opts) }

		Context("scanning for problems failed", func() {
			BeforeEach(func() {
				deployment.ScanForProblemsReturns(nil, errors.New("fake-err"))
			})

			It("returns error", func() {
				err := act()
				Expect(err).To(MatchError(ContainSubstring("fake-err")))
			})
		})

		Context("no problems are found", func() {
			BeforeEach(func() {
				deployment.ScanForProblemsReturns([]boshdir.Problem{}, nil)
			})

			It("tells the user", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Said).To(ContainElement("No problems found\n"))
			})

			It("does not ask for confirmation or with choices", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.AskedChoiceCalled).To(BeFalse())
				Expect(ui.AskedConfirmationCalled).To(BeFalse())
			})

			It("does not write a file", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeFS.WriteFileCallCount).To(BeZero())
			})
		})

		Context("problems are found", func() {
			BeforeEach(func() {
				deployment.ScanForProblemsReturns(severalProbs, nil)
				ui.AskedChoiceChosens = []int{0, 1, 2}
				ui.AskedChoiceErrs = []error{nil, nil, nil}
				ui.AskedConfirmationErr = nil
				ui.AskedText = []fakeui.Answer{
					{Text: "10", Error: nil},
					{Text: "50%", Error: nil},
				}
			})

			It("shows problems by instance group and type", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Said).To(ContainElements(
					"Instance Group 'router'\n",
					"Instance Group 'diego_cell'\n",
				))
				Expect(ui.Tables).To(
					ContainElements(
						boshtbl.Table{
							Title:   "Problem type: unresponsive_agent",
							Content: "unresponsive_agent problems",

							Header: []boshtbl.Header{
								boshtbl.NewHeader("#"),
								boshtbl.NewHeader("Description"),
							},

							SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

							Rows: [][]boshtbl.Value{
								{
									boshtbl.NewValueInt(3),
									boshtbl.NewValueString("problem1-desc"),
								},
							},
						},
						boshtbl.Table{
							Title:   "Problem type: missing_vm",
							Content: "missing_vm problems",

							Header: []boshtbl.Header{
								boshtbl.NewHeader("#"),
								boshtbl.NewHeader("Description"),
							},

							SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

							Rows: [][]boshtbl.Value{
								{
									boshtbl.NewValueInt(4),
									boshtbl.NewValueString("problem2-desc"),
								},
							},
						},
						boshtbl.Table{
							Title:   "Problem type: mount_info_mismatch",
							Content: "mount_info_mismatch problems",

							Header: []boshtbl.Header{
								boshtbl.NewHeader("#"),
								boshtbl.NewHeader("Description"),
							},

							SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

							Rows: [][]boshtbl.Value{
								{
									boshtbl.NewValueInt(5),
									boshtbl.NewValueString("problem3-desc"),
								},
							},
						},
					),
				)
			})

			It("writes a recovery plan based on answers", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.AskedChoiceCalled).To(BeTrue())

				Expect(fakeFS.WriteFileCallCount).To(Equal(1))
				Expect(fakeFS.FileExists("/tmp/foo.yml")).To(BeTrue())
				bytes, err := fakeFS.ReadFile("/tmp/foo.yml")
				Expect(err).ToNot(HaveOccurred())

				var actualPlan RecoveryPlan
				Expect(yaml.Unmarshal(bytes, &actualPlan)).ToNot(HaveOccurred())

				Expect(actualPlan.InstanceGroupsPlan).To(HaveLen(2))

				Expect(actualPlan.InstanceGroupsPlan[0].Name).To(Equal("diego_cell"))
				Expect(actualPlan.InstanceGroupsPlan[0].MaxInFlightOverride).To(Equal("10"))
				Expect(actualPlan.InstanceGroupsPlan[0].PlannedResolutions).To(HaveLen(1))
				Expect(actualPlan.InstanceGroupsPlan[0].PlannedResolutions).To(HaveKeyWithValue("unresponsive_agent", *skipResolution.Name))

				Expect(actualPlan.InstanceGroupsPlan[1].Name).To(Equal("router"))
				Expect(actualPlan.InstanceGroupsPlan[1].MaxInFlightOverride).To(Equal("50%"))
				Expect(actualPlan.InstanceGroupsPlan[1].PlannedResolutions).To(HaveLen(2))
				Expect(actualPlan.InstanceGroupsPlan[1].PlannedResolutions).To(HaveKeyWithValue("missing_vm", *recreateResolution.Name))
				Expect(actualPlan.InstanceGroupsPlan[1].PlannedResolutions).To(HaveKeyWithValue("mount_info_mismatch", *reattachDiskAndRebootResolution.Name))
			})

			It("returns an error if asking fails", func() {
				ui.AskedChoiceErrs = []error{nil, errors.New("fake-err"), nil}

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})

			It("does not override max_in_flight if not confirmed", func() {
				ui.AskedConfirmationErr = errors.New("fake-err")

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeFS.WriteFileCallCount).To(Equal(1))
				Expect(fakeFS.FileExists("/tmp/foo.yml")).To(BeTrue())
				bytes, err := fakeFS.ReadFile("/tmp/foo.yml")
				Expect(err).ToNot(HaveOccurred())

				var actualPlan RecoveryPlan
				Expect(yaml.Unmarshal(bytes, &actualPlan)).ToNot(HaveOccurred())

				Expect(actualPlan.InstanceGroupsPlan).To(HaveLen(2))

				Expect(actualPlan.InstanceGroupsPlan[0].Name).To(Equal("diego_cell"))
				Expect(actualPlan.InstanceGroupsPlan[0].MaxInFlightOverride).To(BeEmpty())

				Expect(actualPlan.InstanceGroupsPlan[1].Name).To(Equal("router"))
				Expect(actualPlan.InstanceGroupsPlan[1].MaxInFlightOverride).To(BeEmpty())
			})

			Context("director does not return instance group", func() {
				BeforeEach(func() {
					grouplessProbs := severalProbs
					for i := range grouplessProbs {
						grouplessProbs[i].InstanceGroup = ""
					}

					deployment.ScanForProblemsReturns(grouplessProbs, nil)
				})

				It("informs the user and exits", func() {
					err := act()
					Expect(err).To(MatchError(ContainSubstring("does not support")))
				})
			})
		})
	})
})
