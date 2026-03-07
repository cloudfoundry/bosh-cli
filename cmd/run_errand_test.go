package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	fakecmd "github.com/cloudfoundry/bosh-cli/v7/cmd/cmdfakes"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	fakessh "github.com/cloudfoundry/bosh-cli/v7/ssh/sshfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

var _ = Describe("RunErrandCmd", func() {
	var (
		deployment *fakedir.FakeDeployment
		downloader *fakecmd.FakeDownloader
		ui         *fakeui.FakeUI
		scpRunner  *fakessh.FakeSCPRunner
		fs         *fakesys.FakeFileSystem
		logger     boshlog.Logger
		command    cmd.RunErrandCmd
	)

	BeforeEach(func() {
		deployment = &fakedir.FakeDeployment{}
		downloader = &fakecmd.FakeDownloader{}
		ui = &fakeui.FakeUI{}
		scpRunner = &fakessh.FakeSCPRunner{}
		fs = fakesys.NewFakeFileSystem()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		command = cmd.NewRunErrandCmd(deployment, downloader, ui, scpRunner, fs, logger)
	})

	Describe("Run", func() {
		var (
			runErrandOpts opts.RunErrandOpts
		)

		BeforeEach(func() {
			runErrandOpts = opts.RunErrandOpts{
				Args:        opts.RunErrandArgs{Name: "errand-name"},
				KeepAlive:   true,
				WhenChanged: true,
				InstanceGroupOrInstanceSlugFlags: opts.InstanceGroupOrInstanceSlugFlags{
					Slugs: []boshdir.InstanceGroupOrInstanceSlug{
						boshdir.NewInstanceGroupOrInstanceSlug("group2", "uuid"),
					},
				},
			}
		})

		act := func() error { return command.Run(runErrandOpts) }

		Context("when errand succeeds", func() {
			Context("when multiple errands return", func() {
				It("downloads logs if requested", func() {
					runErrandOpts.DownloadLogs = true
					runErrandOpts.LogsDirectory = opts.DirOrCWDArg{Path: "/fake-dir"}

					result := []boshdir.ErrandResult{{
						ExitCode:        0,
						LogsBlobstoreID: "logs-blob-id",
						LogsSHA1:        "logs-sha1",
					}, {
						ExitCode:        0,
						LogsBlobstoreID: "logs-blob-id2",
						LogsSHA1:        "logs-sha2",
					}}

					deployment.RunErrandReturns(result, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(downloader.DownloadCallCount()).To(Equal(2))

					blobID, sha1, prefix, dstDirPath := downloader.DownloadArgsForCall(0)
					Expect(blobID).To(Equal("logs-blob-id"))
					Expect(sha1).To(Equal("logs-sha1"))
					Expect(prefix).To(Equal("errand-name"))
					Expect(dstDirPath).To(Equal("/fake-dir"))

					blobID, sha1, prefix, dstDirPath = downloader.DownloadArgsForCall(1)
					Expect(blobID).To(Equal("logs-blob-id2"))
					Expect(sha1).To(Equal("logs-sha2"))
					Expect(prefix).To(Equal("errand-name"))
					Expect(dstDirPath).To(Equal("/fake-dir"))
				})

				It("does not download logs if not requested", func() {
					runErrandOpts.DownloadLogs = false

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(downloader.DownloadCallCount()).To(Equal(0))
				})

				It("does not download logs if requested and not logs blob returned", func() {
					runErrandOpts.DownloadLogs = true
					runErrandOpts.LogsDirectory = opts.DirOrCWDArg{Path: "/fake-dir"}

					result := []boshdir.ErrandResult{{ExitCode: 0}}

					deployment.RunErrandReturns(result, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(downloader.DownloadCallCount()).To(Equal(0))
				})

				It("runs errand and outputs both stdout and stderr", func() {
					result := []boshdir.ErrandResult{{
						InstanceGroup: "group1",
						InstanceID:    "uuid-1",
						ExitCode:      0,
						Stdout:        "stdout-content",
						Stderr:        "",
					}, {
						ExitCode: 129,
						Stdout:   "",
						Stderr:   "stderr-content",
					}}

					deployment.RunErrandReturns(result, nil)

					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Errand 'errand-name' was canceled (exit code 129)"))

					Expect(ui.Table).To(Equal(
						boshtbl.Table{
							Content: "errand(s)",

							Header: []boshtbl.Header{
								boshtbl.NewHeader("Instance"),
								boshtbl.NewHeader("Exit Code"),
								boshtbl.NewHeader("Stdout"),
								boshtbl.NewHeader("Stderr"),
							},

							SortBy: []boshtbl.ColumnSort{
								{Column: 0, Asc: true},
							},

							Rows: [][]boshtbl.Value{
								{
									boshtbl.NewValueString("group1/uuid-1"),
									boshtbl.NewValueInt(0),
									boshtbl.NewValueString("stdout-content"),
									boshtbl.NewValueString(""),
								}, {
									boshtbl.NewValueString(""),
									boshtbl.NewValueInt(129),
									boshtbl.NewValueString(""),
									boshtbl.NewValueString("stderr-content"),
								},
							},

							Notes: []string{},

							FillFirstColumn: true,

							Transpose: true,
						}))
				})

			})

			It("runs errand with given name", func() {
				deployment.RunErrandReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)
				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.RunErrandCallCount()).To(Equal(1))

				name, keepAlive, whenChanged, slugs := deployment.RunErrandArgsForCall(0)
				Expect(name).To(Equal("errand-name"))
				Expect(keepAlive).To(BeTrue())
				Expect(whenChanged).To(BeTrue())
				Expect(slugs).To(HaveLen(1))
				Expect(slugs[0].Name()).To(Equal("group2"))
				Expect(slugs[0].IndexOrID()).To(Equal("uuid"))
			})

			It("downloads logs if requested", func() {
				runErrandOpts.DownloadLogs = true
				runErrandOpts.LogsDirectory = opts.DirOrCWDArg{Path: "/fake-dir"}

				result := []boshdir.ErrandResult{{
					ExitCode:        0,
					LogsBlobstoreID: "logs-blob-id",
					LogsSHA1:        "logs-sha1",
				}}

				deployment.RunErrandReturns(result, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(downloader.DownloadCallCount()).To(Equal(1))

				blobID, sha1, prefix, dstDirPath := downloader.DownloadArgsForCall(0)
				Expect(blobID).To(Equal("logs-blob-id"))
				Expect(sha1).To(Equal("logs-sha1"))
				Expect(prefix).To(Equal("errand-name"))
				Expect(dstDirPath).To(Equal("/fake-dir"))
			})

			It("does not download logs if not requested", func() {
				runErrandOpts.DownloadLogs = false

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(downloader.DownloadCallCount()).To(Equal(0))
			})

			It("does not download logs if requested and not logs blob returned", func() {
				runErrandOpts.DownloadLogs = true
				runErrandOpts.LogsDirectory = opts.DirOrCWDArg{Path: "/fake-dir"}

				result := []boshdir.ErrandResult{{ExitCode: 0}}

				deployment.RunErrandReturns(result, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(downloader.DownloadCallCount()).To(Equal(0))
			})
		})

		Context("when errand fails (exit code is non-0)", func() {
			It("returns error", func() {
				deployment.RunErrandReturns([]boshdir.ErrandResult{{ExitCode: 1}}, nil)

				err := act()
				Expect(err).To(HaveOccurred())

				Expect(ui.Table).To(Equal(
					boshtbl.Table{
						Content: "errand(s)",

						Header: []boshtbl.Header{
							boshtbl.NewHeader("Instance"),
							boshtbl.NewHeader("Exit Code"),
							boshtbl.NewHeader("Stdout"),
							boshtbl.NewHeader("Stderr"),
						},

						SortBy: []boshtbl.ColumnSort{
							{Column: 0, Asc: true},
						},

						Rows: [][]boshtbl.Value{
							{
								boshtbl.NewValueString(""),
								boshtbl.NewValueInt(1),
								boshtbl.NewValueString(""),
								boshtbl.NewValueString(""),
							},
						},

						Notes: []string{},

						FillFirstColumn: true,

						Transpose: true,
					}))

				Expect(err.Error()).To(Equal("Errand 'errand-name' completed with error (exit code 1)"))
			})
		})

		Context("when errand is canceled (exit code > 128)", func() {
			It("returns error", func() {
				deployment.RunErrandReturns([]boshdir.ErrandResult{{ExitCode: 129}}, nil)

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Errand 'errand-name' was canceled (exit code 129)"))
			})
		})

		It("returns error if running errand failed", func() {
			deployment.RunErrandReturns([]boshdir.ErrandResult{{}}, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})

	Describe("Run with --stream-logs", func() {
		var (
			runErrandOpts opts.RunErrandOpts
			uuidGen       *fakeuuid.FakeGenerator
		)

		BeforeEach(func() {
			uuidGen = fakeuuid.NewFakeGenerator()
			uuidGen.GeneratedUUID = "8c5ff117-9572-45c5-8564-8bcf076ecafa"
			streamInterval := 30
			runErrandOpts = opts.RunErrandOpts{
				Args:         opts.RunErrandArgs{Name: "my-errand"},
				StreamLogs:   &streamInterval,
				GatewayFlags: opts.GatewayFlags{UUIDGen: uuidGen},
			}
		})

		act := func() error { return command.Run(runErrandOpts) }

		It("uses normal path when StreamLogs is nil", func() {
			runErrandOpts.StreamLogs = nil
			deployment.RunErrandReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

			err := command.Run(runErrandOpts)
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.RunErrandCallCount()).To(Equal(1))
			Expect(deployment.StartErrandCallCount()).To(Equal(0))
		})

		Context("with --instance flag", func() {
			BeforeEach(func() {
				runErrandOpts.InstanceGroupOrInstanceSlugFlags = opts.InstanceGroupOrInstanceSlugFlags{
					Slugs: []boshdir.InstanceGroupOrInstanceSlug{
						boshdir.NewInstanceGroupOrInstanceSlug("web_server", ""),
					},
				}
			})

			Context("when SSH setup fails", func() {
				It("falls back to non-streaming path", func() {
					deployment.SetUpSSHReturns(boshdir.SSHResult{}, errors.New("ssh-setup-err"))
					deployment.RunErrandReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(deployment.RunErrandCallCount()).To(Equal(1))
					Expect(deployment.StartErrandCallCount()).To(Equal(0))
				})
			})

			Context("when SSH setup succeeds", func() {
				BeforeEach(func() {
					deployment.SetUpSSHReturns(boshdir.SSHResult{
						Hosts: []boshdir.Host{
							{Job: "web_server", IndexOrID: "abc-123", Username: "vcap", Host: "10.0.0.1"},
						},
					}, nil)
				})

				It("sets up SSH using the instance group from --instance", func() {
					deployment.StartErrandReturns(42, nil)
					deployment.WaitForErrandSilentlyReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(deployment.SetUpSSHCallCount()).To(Equal(1))
					slug, _ := deployment.SetUpSSHArgsForCall(0)
					Expect(slug.Name()).To(Equal("web_server"))

					Expect(deployment.ManifestCallCount()).To(Equal(0))
				})

				It("passes instance index/ID through to SetUpSSH slug", func() {
					runErrandOpts.InstanceGroupOrInstanceSlugFlags = opts.InstanceGroupOrInstanceSlugFlags{
						Slugs: []boshdir.InstanceGroupOrInstanceSlug{
							boshdir.NewInstanceGroupOrInstanceSlug("web_server", "0"),
						},
					}
					deployment.StartErrandReturns(42, nil)
					deployment.WaitForErrandSilentlyReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(deployment.SetUpSSHCallCount()).To(Equal(1))
					slug, _ := deployment.SetUpSSHArgsForCall(0)
					Expect(slug.Name()).To(Equal("web_server"))
					Expect(slug.IndexOrID()).To(Equal("0"))
				})

				It("starts errand, waits for it silently, and summarizes results", func() {
					deployment.StartErrandReturns(42, nil)
					deployment.WaitForErrandSilentlyReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(deployment.StartErrandCallCount()).To(Equal(1))
					name, keepAlive, whenChanged, slugs := deployment.StartErrandArgsForCall(0)
					Expect(name).To(Equal("my-errand"))
					Expect(keepAlive).To(BeFalse())
					Expect(whenChanged).To(BeFalse())
					Expect(slugs).To(HaveLen(1))
					Expect(slugs[0].Name()).To(Equal("web_server"))

					Expect(deployment.WaitForErrandSilentlyCallCount()).To(Equal(1))
					Expect(deployment.WaitForErrandSilentlyArgsForCall(0)).To(Equal(42))

					Expect(deployment.CleanUpSSHCallCount()).To(Equal(1))
				})

				It("replays task events via WaitForErrand after final flush", func() {
					deployment.StartErrandReturns(42, nil)
					deployment.WaitForErrandSilentlyReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(deployment.WaitForErrandCallCount()).To(Equal(1))
					Expect(deployment.WaitForErrandArgsForCall(0)).To(Equal(42))
				})

				It("shows streaming message with instance list", func() {
					deployment.StartErrandReturns(42, nil)
					deployment.WaitForErrandSilentlyReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(ui.Blocks).To(ContainElement(ContainSubstring("Streaming logs for errand 'my-errand'")))
					Expect(ui.Blocks).To(ContainElement(ContainSubstring("web_server/abc-123")))
				})

				It("shows stdout as streamed in summary table", func() {
					deployment.StartErrandReturns(42, nil)
					deployment.WaitForErrandSilentlyReturns([]boshdir.ErrandResult{{
						InstanceGroup: "web_server",
						InstanceID:    "abc-123",
						ExitCode:      0,
						Stdout:        "real stdout content",
						Stderr:        "some stderr",
					}}, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(ui.Table.Rows).To(HaveLen(1))
					Expect(ui.Table.Rows[0][2]).To(Equal(boshtbl.NewValueString("(errand output was streamed above)")))
					Expect(ui.Table.Rows[0][3]).To(Equal(boshtbl.NewValueString("some stderr")))
				})

				It("returns error when StartErrand fails", func() {
					deployment.StartErrandReturns(0, errors.New("start-err"))

					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("start-err"))

					Expect(deployment.CleanUpSSHCallCount()).To(Equal(1))
				})

				It("returns error when WaitForErrandSilently fails", func() {
					deployment.StartErrandReturns(42, nil)
					deployment.WaitForErrandSilentlyReturns(nil, errors.New("wait-err"))

					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("wait-err"))
				})

				It("returns errand error for non-zero exit code", func() {
					deployment.StartErrandReturns(42, nil)
					deployment.WaitForErrandSilentlyReturns([]boshdir.ErrandResult{{ExitCode: 1}}, nil)

					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("Errand 'my-errand' completed with error (exit code 1)"))
				})
			})
		})

		Context("with custom poll interval", func() {
			It("accepts a custom poll interval via --stream-logs=N", func() {
				customInterval := 10
				runErrandOpts.StreamLogs = &customInterval
				runErrandOpts.InstanceGroupOrInstanceSlugFlags = opts.InstanceGroupOrInstanceSlugFlags{
					Slugs: []boshdir.InstanceGroupOrInstanceSlug{
						boshdir.NewInstanceGroupOrInstanceSlug("web_server", ""),
					},
				}
				deployment.SetUpSSHReturns(boshdir.SSHResult{
					Hosts: []boshdir.Host{
						{Job: "web_server", IndexOrID: "abc-123", Username: "vcap", Host: "10.0.0.1"},
					},
				}, nil)
				deployment.StartErrandReturns(42, nil)
				deployment.WaitForErrandSilentlyReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Blocks).To(ContainElement(ContainSubstring("poll interval 10s")))
			})
		})

		Context("without --instance flag (manifest lookup)", func() {
			It("looks up the instance group from the manifest and sets up SSH", func() {
				deployment.ManifestReturns("instance_groups:\n- name: web_server\n  jobs:\n  - name: my-errand\n    release: my-release\n", nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{
					Hosts: []boshdir.Host{
						{Job: "web_server", IndexOrID: "abc-def-123", Username: "vcap", Host: "10.0.0.1"},
					},
				}, nil)
				deployment.StartErrandReturns(42, nil)
				deployment.WaitForErrandSilentlyReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.ManifestCallCount()).To(Equal(1))

				Expect(deployment.SetUpSSHCallCount()).To(Equal(1))
				slug, _ := deployment.SetUpSSHArgsForCall(0)
				Expect(slug.Name()).To(Equal("web_server"))

				Expect(deployment.StartErrandCallCount()).To(Equal(1))
				name, _, _, slugs := deployment.StartErrandArgsForCall(0)
				Expect(name).To(Equal("my-errand"))
				Expect(slugs).To(BeNil())

				Expect(deployment.WaitForErrandSilentlyCallCount()).To(Equal(1))
				Expect(deployment.CleanUpSSHCallCount()).To(Equal(1))
			})

			It("falls back to non-streaming when manifest fetch fails", func() {
				deployment.ManifestReturns("", errors.New("manifest-err"))
				deployment.RunErrandReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.RunErrandCallCount()).To(Equal(1))
				Expect(deployment.StartErrandCallCount()).To(Equal(0))
			})

			It("falls back to non-streaming when errand job not found in manifest", func() {
				deployment.ManifestReturns("instance_groups:\n- name: web_server\n  jobs:\n  - name: other-job\n    release: my-release\n", nil)
				deployment.RunErrandReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.RunErrandCallCount()).To(Equal(1))
				Expect(deployment.StartErrandCallCount()).To(Equal(0))
			})

			It("falls back to non-streaming when SSH setup fails", func() {
				deployment.ManifestReturns("instance_groups:\n- name: web_server\n  jobs:\n  - name: my-errand\n    release: my-release\n", nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{}, errors.New("ssh-fail"))
				deployment.RunErrandReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.RunErrandCallCount()).To(Equal(1))
				Expect(deployment.StartErrandCallCount()).To(Equal(0))
			})

			It("returns error when StartErrand fails", func() {
				deployment.ManifestReturns("instance_groups:\n- name: web_server\n  jobs:\n  - name: my-errand\n    release: my-release\n", nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{
					Hosts: []boshdir.Host{
						{Job: "web_server", IndexOrID: "abc-def-123", Username: "vcap", Host: "10.0.0.1"},
					},
				}, nil)
				deployment.StartErrandReturns(0, errors.New("start-err"))

				err := act()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("start-err"))
			})
		})
	})
})
