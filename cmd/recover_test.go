package cmd_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

var _ = Describe("RecoverCmd", func() {
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
		command    cmd.RecoverCmd
	)

	BeforeEach(func() {
		deployment = &fakedir.FakeDeployment{}
		ui = &fakeui.FakeUI{}
		fakeFS = fakesys.NewFakeFileSystem()
		command = cmd.NewRecoverCmd(deployment, ui, fakeFS)
	})

	Describe("Run", func() {
		var (
			recoverOpts  opts.RecoverOpts
			severalProbs []boshdir.Problem
			plan         cmd.RecoveryPlan
		)

		BeforeEach(func() {
			recoverOpts = opts.RecoverOpts{
				Args: opts.RecoverArgs{
					RecoveryPlan: opts.FileArg{
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

			plan = cmd.RecoveryPlan{
				InstanceGroupsPlan: []cmd.InstanceGroupPlan{
					{
						Name:                "diego_cell",
						MaxInFlightOverride: "10",
						PlannedResolutions: map[string]string{
							"unresponsive_agent": *skipResolution.Name,
						},
					},
					{
						Name: "router",
						PlannedResolutions: map[string]string{
							"missing_vm":          *recreateResolution.Name,
							"mount_info_mismatch": *reattachDiskAndRebootResolution.Name,
						},
					},
				},
			}

			bytes, err := yaml.Marshal(plan)
			Expect(err).NotTo(HaveOccurred())

			fakeFile := fakesys.NewFakeFile("/tmp/foo.yml", fakeFS)
			_, err = fakeFile.Write(bytes)
			Expect(err).NotTo(HaveOccurred())
		})

		act := func() error { return command.Run(recoverOpts) }

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

			It("does not ask for confirmation", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.AskedConfirmationCalled).To(BeFalse())
			})
		})

		Context("problems are found", func() {
			BeforeEach(func() {
				deployment.ScanForProblemsReturns(severalProbs, nil)
				ui.AskedConfirmationErr = nil
			})

			It("returns an error if reading recovery plan fails", func() {
				fakeFS.ReadFileError = errors.New("fake-err")

				err := act()
				Expect(err).To(MatchError("fake-err"))
			})

			Context("applying recovery plan", func() {
				It("displays a plan summary", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(ui.Tables).To(
						ContainElements(
							boshtbl.Table{
								Title: "Instance Group 'diego_cell' plan summary (max_in_flight override: 10)",

								Header: []boshtbl.Header{
									boshtbl.NewHeader("#"),
									boshtbl.NewHeader("Planned resolution"),
									boshtbl.NewHeader("Description"),
								},

								SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueInt(3),
										boshtbl.NewValueString(skipResolution.Plan),
										boshtbl.NewValueString("problem1-desc"),
									},
								},
							},
							boshtbl.Table{
								Title: "Instance Group 'router' plan summary",

								Header: []boshtbl.Header{
									boshtbl.NewHeader("#"),
									boshtbl.NewHeader("Planned resolution"),
									boshtbl.NewHeader("Description"),
								},

								SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

								Rows: [][]boshtbl.Value{
									{
										boshtbl.NewValueInt(4),
										boshtbl.NewValueString(recreateResolution.Plan),
										boshtbl.NewValueString("problem2-desc"),
									},
									{
										boshtbl.NewValueInt(5),
										boshtbl.NewValueString(reattachDiskAndRebootResolution.Plan),
										boshtbl.NewValueString("problem3-desc"),
									},
								},
							},
						),
					)
				})

				It("asks for confirmation", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(ui.AskedConfirmationCalled).To(BeTrue())
				})

				Context("not confirmed", func() {
					BeforeEach(func() {
						ui.AskedConfirmationErr = errors.New("nope")
					})

					It("does not call deployment func", func() {
						err := act()
						Expect(err).To(MatchError("nope"))

						Expect(deployment.ResolveProblemsCallCount()).To(Equal(0))
					})
				})

				Context("confirmed", func() {
					It("calls the appropriate deployment func", func() {
						err := act()
						Expect(err).ToNot(HaveOccurred())

						Expect(deployment.ResolveProblemsCallCount()).To(Equal(1))

						answers, overrides := deployment.ResolveProblemsArgsForCall(0)
						Expect(answers).To(ContainElements(
							boshdir.ProblemAnswer{ProblemID: 3, Resolution: skipResolution},
							boshdir.ProblemAnswer{ProblemID: 4, Resolution: recreateResolution},
							boshdir.ProblemAnswer{ProblemID: 5, Resolution: reattachDiskAndRebootResolution},
						))
						Expect(overrides).To(Equal(map[string]string{
							"diego_cell": "10",
						}))
					})
				})
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
