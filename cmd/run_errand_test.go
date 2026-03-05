package cmd_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	fakecmd "github.com/cloudfoundry/bosh-cli/v7/cmd/cmdfakes"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	boshssh "github.com/cloudfoundry/bosh-cli/v7/ssh"
	fakessh "github.com/cloudfoundry/bosh-cli/v7/ssh/sshfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
)

var _ = Describe("RunErrandCmd", func() {
	var (
		deployment *fakedir.FakeDeployment
		downloader *fakecmd.FakeDownloader
		ui         *fakeui.FakeUI
		command    cmd.RunErrandCmd
	)

	BeforeEach(func() {
		deployment = &fakedir.FakeDeployment{}
		downloader = &fakecmd.FakeDownloader{}
		ui = &fakeui.FakeUI{}
		command = cmd.NewRunErrandCmd(deployment, downloader, ui, nil, nil, nil)
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

		Context("when --stream-logs is NOT set", func() {
			It("does not set up SSH or run SSH runner", func() {
				deployment.RunErrandReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

				err := act()
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.SetUpSSHCallCount()).To(Equal(0))
				Expect(deployment.CleanUpSSHCallCount()).To(Equal(0))
				Expect(deployment.StartErrandCallCount()).To(Equal(0))
				Expect(deployment.RunErrandCallCount()).To(Equal(1))
			})
		})

		Context("when --stream-logs is set", func() {
			const UUID = "8c5ff117-9572-45c5-8564-8bcf076ecafa"

			var (
				nonIntSSHRunner *fakessh.FakeRunner
				uuidGen         *fakeuuid.FakeGenerator
				streamOpts      opts.RunErrandOpts
			)

			makeEvent := func(stage, state, task string) string {
				ev := map[string]any{
					"stage": stage,
					"state": state,
					"task":  task,
					"time":  1772657703,
				}
				b, err := json.Marshal(ev)
				Expect(err).NotTo(HaveOccurred())
				return string(b)
			}

			BeforeEach(func() {
				nonIntSSHRunner = &fakessh.FakeRunner{}
				uuidGen = &fakeuuid.FakeGenerator{}
				uuidGen.GeneratedUUID = UUID

				command = cmd.NewRunErrandCmd(deployment, downloader, ui, nonIntSSHRunner, nil, nil)

				streamOpts = opts.RunErrandOpts{
					Args:        opts.RunErrandArgs{Name: "smoke-tests"},
					StreamLogs:  true,
					KeepAlive:   true,
					WhenChanged: false,
				}
				streamOpts.GatewayFlags.UUIDGen = uuidGen //nolint:staticcheck
			})

			It("calls StartErrand instead of RunErrand", func() {
				events := makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)")
				deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
					return []byte(events), len(events), nil
				}
				deployment.TaskStateReturns("done", nil)
				deployment.StartErrandReturns(42, nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{
					Hosts: []boshdir.Host{{Host: "10.0.0.1", Job: "smoke-tests", IndexOrID: "abc-123"}},
				}, nil)
				deployment.WaitForErrandResultReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

				err := command.Run(streamOpts)
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.RunErrandCallCount()).To(Equal(0))
				Expect(deployment.StartErrandCallCount()).To(Equal(1))

				name, keepAlive, whenChanged, slugs := deployment.StartErrandArgsForCall(0)
				Expect(name).To(Equal("smoke-tests"))
				Expect(keepAlive).To(BeTrue())
				Expect(whenChanged).To(BeFalse())
				Expect(slugs).To(BeEmpty())
			})

			It("sets up and tears down SSH for discovered instances", func() {
				events := makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)")
				deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
					return []byte(events), len(events), nil
				}
				deployment.TaskStateReturns("done", nil)
				deployment.StartErrandReturns(42, nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{
					Hosts: []boshdir.Host{{Host: "10.0.0.1", Job: "smoke-tests", IndexOrID: "abc-123"}},
				}, nil)
				deployment.WaitForErrandResultReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

				err := command.Run(streamOpts)
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.SetUpSSHCallCount()).To(BeNumerically(">=", 1))
				Expect(deployment.CleanUpSSHCallCount()).To(BeNumerically(">=", 1))

				setupSlug, _ := deployment.SetUpSSHArgsForCall(0)
				Expect(setupSlug).To(Equal(boshdir.NewAllOrInstanceGroupOrInstanceSlug("smoke-tests", "abc-123")))
			})

			It("runs tail command with correct default path", func() {
				events := makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)")
				deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
					return []byte(events), len(events), nil
				}
				deployment.TaskStateReturns("done", nil)
				deployment.StartErrandReturns(42, nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{
					Hosts: []boshdir.Host{{Host: "10.0.0.1", Job: "smoke-tests", IndexOrID: "abc-123"}},
				}, nil)
				deployment.WaitForErrandResultReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

				err := command.Run(streamOpts)
				Expect(err).ToNot(HaveOccurred())

				Expect(nonIntSSHRunner.RunContextCallCount()).To(BeNumerically(">=", 1))

				_, _, _, runCmd := nonIntSSHRunner.RunContextArgsForCall(0)
				Expect(runCmd).To(HaveLen(4))
				Expect(runCmd[0]).To(Equal("sudo"))
				Expect(runCmd[1]).To(Equal("bash"))
				Expect(runCmd[2]).To(Equal("-c"))
				Expect(runCmd[3]).To(ContainSubstring("tail -n 0 -F"))
				Expect(runCmd[3]).To(ContainSubstring("/var/vcap/sys/log/smoke-tests/smoke-tests.{stdout,stderr}.log"))
			})

			It("uses custom --stream-log-path when specified", func() {
				streamOpts.StreamLogPath = "my-job/custom.log"

				events := makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)")
				deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
					return []byte(events), len(events), nil
				}
				deployment.TaskStateReturns("done", nil)
				deployment.StartErrandReturns(42, nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{
					Hosts: []boshdir.Host{{Host: "10.0.0.1", Job: "smoke-tests", IndexOrID: "abc-123"}},
				}, nil)
				deployment.WaitForErrandResultReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

				err := command.Run(streamOpts)
				Expect(err).ToNot(HaveOccurred())

				Expect(nonIntSSHRunner.RunContextCallCount()).To(BeNumerically(">=", 1))
				_, _, _, runCmd := nonIntSSHRunner.RunContextArgsForCall(0)
				Expect(runCmd[3]).To(ContainSubstring("/var/vcap/sys/log/my-job/custom.log"))
			})

			It("passes gateway flags through to SSH connection", func() {
				streamOpts.GatewayFlags.Disable = true                    //nolint:staticcheck
				streamOpts.GatewayFlags.Username = "gw-user"              //nolint:staticcheck
				streamOpts.GatewayFlags.Host = "gw-host"                  //nolint:staticcheck
				streamOpts.GatewayFlags.PrivateKeyPath = "gw-private-key" //nolint:staticcheck
				streamOpts.GatewayFlags.SOCKS5Proxy = "some-proxy"        //nolint:staticcheck

				events := makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)")
				deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
					return []byte(events), len(events), nil
				}
				deployment.TaskStateReturns("done", nil)
				deployment.StartErrandReturns(42, nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{
					Hosts: []boshdir.Host{{Host: "10.0.0.1", Job: "smoke-tests", IndexOrID: "abc-123"}},
				}, nil)
				deployment.WaitForErrandResultReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

				err := command.Run(streamOpts)
				Expect(err).ToNot(HaveOccurred())

				Expect(nonIntSSHRunner.RunContextCallCount()).To(BeNumerically(">=", 1))
				_, connOpts, _, _ := nonIntSSHRunner.RunContextArgsForCall(0)
				Expect(connOpts.GatewayDisable).To(BeTrue())
				Expect(connOpts.GatewayUsername).To(Equal("gw-user"))
				Expect(connOpts.GatewayHost).To(Equal("gw-host"))
				Expect(connOpts.GatewayPrivateKeyPath).To(Equal("gw-private-key"))
				Expect(connOpts.SOCKS5Proxy).To(Equal("some-proxy"))
			})

			It("still prints errand result summary after streaming", func() {
				events := makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)")
				deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
					return []byte(events), len(events), nil
				}
				deployment.TaskStateReturns("done", nil)
				deployment.StartErrandReturns(42, nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{
					Hosts: []boshdir.Host{{Host: "10.0.0.1", Job: "smoke-tests", IndexOrID: "abc-123"}},
				}, nil)
				deployment.WaitForErrandResultReturns([]boshdir.ErrandResult{{
					InstanceGroup: "smoke-tests",
					InstanceID:    "abc-123",
					ExitCode:      0,
					Stdout:        "test output",
				}}, nil)

				err := command.Run(streamOpts)
				Expect(err).ToNot(HaveOccurred())

				Expect(ui.Table.Content).To(Equal("errand(s)"))
				Expect(ui.Table.Rows).To(HaveLen(1))
				Expect(ui.Table.Rows[0][0]).To(Equal(boshtbl.NewValueString("smoke-tests/abc-123")))
			})

			It("handles errand failure with exit code", func() {
				events := makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)")
				deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
					return []byte(events), len(events), nil
				}
				deployment.TaskStateReturns("done", nil)
				deployment.StartErrandReturns(42, nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{
					Hosts: []boshdir.Host{{Host: "10.0.0.1", Job: "smoke-tests", IndexOrID: "abc-123"}},
				}, nil)
				deployment.WaitForErrandResultReturns([]boshdir.ErrandResult{{ExitCode: 1}}, nil)

				err := command.Run(streamOpts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Errand 'smoke-tests' completed with error (exit code 1)"))
			})

			It("returns error when StartErrand fails", func() {
				deployment.StartErrandReturns(0, errors.New("async-err"))

				err := command.Run(streamOpts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("async-err"))
			})

			It("returns error when WaitForErrandResult fails", func() {
				events := makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)")
				deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
					return []byte(events), len(events), nil
				}
				deployment.TaskStateReturns("done", nil)
				deployment.StartErrandReturns(42, nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{
					Hosts: []boshdir.Host{{Host: "10.0.0.1", Job: "smoke-tests", IndexOrID: "abc-123"}},
				}, nil)
				deployment.WaitForErrandResultReturns(nil, errors.New("result-fetch-err"))

				err := command.Run(streamOpts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("result-fetch-err"))
			})

			It("returns error when SSH runner is nil", func() {
				nilRunnerCmd := cmd.NewRunErrandCmd(deployment, downloader, ui, nil, nil, nil)
				err := nilRunnerCmd.Run(streamOpts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("SSH runner is required"))
			})

			It("returns error and does not start errand when --stream-log-path is invalid", func() {
				streamOpts.StreamLogPath = "foo; rm -rf /"

				err := command.Run(streamOpts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid characters"))

				Expect(deployment.StartErrandCallCount()).To(Equal(0),
					"errand should not be started when tail command cannot be built")
			})

			It("returns error when SSH options cannot be generated", func() {
				deployment.StartErrandReturns(42, nil)
				uuidGen.GenerateError = errors.New("uuid-gen-broken")

				err := command.Run(streamOpts)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("uuid-gen-broken"))

				Expect(deployment.SetUpSSHCallCount()).To(Equal(0),
					"should not attempt SSH setup when options fail")
			})

			It("feeds task events to the TaskReporter when one is provided", func() {
				reporter := &fakedir.FakeTaskReporter{}
				command = cmd.NewRunErrandCmd(deployment, downloader, ui, nonIntSSHRunner, reporter, nil)

				events := makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)")
				deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
					return []byte(events), len(events), nil
				}
				deployment.TaskStateReturns("done", nil)
				deployment.StartErrandReturns(42, nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{
					Hosts: []boshdir.Host{{Host: "10.0.0.1", Job: "smoke-tests", IndexOrID: "abc-123"}},
				}, nil)
				deployment.WaitForErrandResultReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

				err := command.Run(streamOpts)
				Expect(err).ToNot(HaveOccurred())

				Expect(reporter.TaskStartedCallCount()).To(BeNumerically(">=", 1))
				taskID := reporter.TaskStartedArgsForCall(0)
				Expect(taskID).To(Equal(42))

				Expect(reporter.TaskOutputChunkCallCount()).To(BeNumerically(">=", 1))
				Expect(reporter.TaskFinishedCallCount()).To(BeNumerically(">=", 1))
			})

			It("skips malformed event data that does not produce a valid slug", func() {
				events := strings.Join([]string{
					makeEvent("Running errand", "started", "no-slash-here"),
					makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)"),
				}, "\n")

				callCount := 0
				deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
					callCount++
					if callCount == 1 {
						return []byte(events), len(events), nil
					}
					return nil, offset, nil
				}
				deployment.TaskStateStub = func(taskID int) (string, error) {
					if callCount >= 1 {
						return "done", nil
					}
					return "processing", nil
				}
				deployment.StartErrandReturns(42, nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{
					Hosts: []boshdir.Host{{Host: "10.0.0.1", Job: "smoke-tests", IndexOrID: "abc-123"}},
				}, nil)
				deployment.WaitForErrandResultReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

				err := command.Run(streamOpts)
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.SetUpSSHCallCount()).To(Equal(1),
					"should only set up SSH for the valid slug, not the malformed one")
			})

			It("handles SetUpSSH failure gracefully", func() {
				events := makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)")
				deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
					return []byte(events), len(events), nil
				}
				deployment.TaskStateReturns("done", nil)
				deployment.StartErrandReturns(42, nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{}, errors.New("ssh-setup-err"))
				deployment.WaitForErrandResultReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

				err := command.Run(streamOpts)
				Expect(err).ToNot(HaveOccurred())

				Expect(nonIntSSHRunner.RunContextCallCount()).To(Equal(0))
			})

			It("handles multiple instances from events", func() {
				events := strings.Join([]string{
					makeEvent("Running errand", "started", "mysql/aaa-111 (0)"),
					makeEvent("Running errand", "started", "mysql/bbb-222 (1)"),
				}, "\n")

				deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
					return []byte(events), len(events), nil
				}
				deployment.TaskStateReturns("done", nil)
				deployment.StartErrandReturns(42, nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{
					Hosts: []boshdir.Host{{Host: "10.0.0.1", Job: "mysql", IndexOrID: "0"}},
				}, nil)
				deployment.WaitForErrandResultReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

				streamOpts.Args.Name = "mysql-errand"
				err := command.Run(streamOpts)
				Expect(err).ToNot(HaveOccurred())

				Expect(deployment.SetUpSSHCallCount()).To(Equal(2))
				Expect(nonIntSSHRunner.RunContextCallCount()).To(Equal(2))
			})

			It("handles SSH runner error without failing the errand", func() {
				events := makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)")
				deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
					return []byte(events), len(events), nil
				}
				deployment.TaskStateReturns("done", nil)
				deployment.StartErrandReturns(42, nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{
					Hosts: []boshdir.Host{{Host: "10.0.0.1", Job: "smoke-tests", IndexOrID: "abc-123"}},
				}, nil)
				nonIntSSHRunner.RunContextReturns(errors.New("signal: killed"))
				deployment.WaitForErrandResultReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

				err := command.Run(streamOpts)
				Expect(err).ToNot(HaveOccurred())
			})

			It("does not hang when SSH runner blocks indefinitely", func() {
				events := makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)")
				deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
					return []byte(events), len(events), nil
				}
				deployment.TaskStateReturns("done", nil)
				deployment.StartErrandReturns(42, nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{
					Hosts: []boshdir.Host{{Host: "10.0.0.1", Job: "smoke-tests", IndexOrID: "abc-123"}},
				}, nil)
				deployment.WaitForErrandResultReturns([]boshdir.ErrandResult{{ExitCode: 0}}, nil)

				// Simulate tail -F: block until the context is cancelled.
				// RunContext should cancel the context when the errand finishes,
				// which unblocks this stub deterministically.
				nonIntSSHRunner.RunContextStub = func(ctx context.Context, _ boshssh.ConnectionOpts, _ boshdir.SSHResult, _ []string) error {
					<-ctx.Done()
					return nil
				}

				done := make(chan error, 1)
				go func() { done <- command.Run(streamOpts) }()

				select {
				case err := <-done:
					Expect(err).ToNot(HaveOccurred())
				case <-time.After(5 * time.Second):
					Fail("Run hung — context cancellation did not unblock the SSH runner")
				}
			})

			It("exits gracefully when interrupted (simulated Ctrl+C)", func() {
				ctx, cancel := context.WithCancel(context.Background())
				command = cmd.NewRunErrandCmd(deployment, downloader, ui, nonIntSSHRunner, nil, func() (context.Context, context.CancelFunc) {
					return ctx, cancel
				})

				sshStarted := make(chan struct{})
				nonIntSSHRunner.RunContextStub = func(ctx context.Context, _ boshssh.ConnectionOpts, _ boshdir.SSHResult, _ []string) error {
					close(sshStarted)
					<-ctx.Done()
					return nil
				}

				events := makeEvent("Running errand", "started", "smoke-tests/abc-123 (0)")
				deployment.FetchTaskOutputChunkStub = func(taskID, offset int, type_ string) ([]byte, int, error) {
					return []byte(events), len(events), nil
				}
				// Keep the task "processing" so the watcher doesn't close slugCh
				// before we get a chance to cancel.
				deployment.TaskStateReturns("processing", nil)
				deployment.StartErrandReturns(42, nil)
				deployment.SetUpSSHReturns(boshdir.SSHResult{
					Hosts: []boshdir.Host{{Host: "10.0.0.1", Job: "smoke-tests", IndexOrID: "abc-123"}},
				}, nil)

				done := make(chan error, 1)
				go func() { done <- command.Run(streamOpts) }()

				// Wait for SSH to be running, then simulate Ctrl+C.
				Eventually(sshStarted, 2*time.Second).Should(BeClosed())
				cancel()

				select {
				case err := <-done:
					Expect(err).ToNot(HaveOccurred())
				case <-time.After(5 * time.Second):
					Fail("Run hung — simulated signal did not cause graceful exit")
				}

				Expect(deployment.WaitForErrandResultCallCount()).To(Equal(0),
					"should NOT wait for errand result after interruption")
				Expect(deployment.CleanUpSSHCallCount()).To(BeNumerically(">=", 1),
					"should still clean up SSH sessions after interruption")

				Expect(ui.Said).To(ContainElement(ContainSubstring("Streaming interrupted")))
				Expect(ui.Said).To(ContainElement(ContainSubstring("bosh task 42")))
			})

		})
	})
})

var _ = Describe("BuildErrandTailCmd", func() {
	It("builds default tail command for errand name", func() {
		tailCmd, err := cmd.BuildErrandTailCmd("smoke-tests", "")
		Expect(err).ToNot(HaveOccurred())
		Expect(tailCmd).To(HaveLen(4))
		Expect(tailCmd[0]).To(Equal("sudo"))
		Expect(tailCmd[1]).To(Equal("bash"))
		Expect(tailCmd[2]).To(Equal("-c"))
		Expect(tailCmd[3]).To(Equal(`'until ls /var/vcap/sys/log/smoke-tests/smoke-tests.{stdout,stderr}.log >/dev/null 2>&1;do sleep 1; done && exec tail -n 0 -F /var/vcap/sys/log/smoke-tests/smoke-tests.{stdout,stderr}.log'`))
	})

	It("uses custom path when specified", func() {
		tailCmd, err := cmd.BuildErrandTailCmd("smoke-tests", "my-job/custom.log")
		Expect(err).ToNot(HaveOccurred())
		Expect(tailCmd[3]).To(ContainSubstring("/var/vcap/sys/log/my-job/custom.log"))
		Expect(tailCmd[3]).ToNot(ContainSubstring("smoke-tests"))
	})

	It("accepts glob and brace expansion characters in custom path", func() {
		tailCmd, err := cmd.BuildErrandTailCmd("smoke-tests", "my-job/*.log")
		Expect(err).ToNot(HaveOccurred())
		Expect(tailCmd[3]).To(ContainSubstring("/var/vcap/sys/log/my-job/*.log"))

		tailCmd, err = cmd.BuildErrandTailCmd("smoke-tests", "my-job/{out,err}.log")
		Expect(err).ToNot(HaveOccurred())
		Expect(tailCmd[3]).To(ContainSubstring("/var/vcap/sys/log/my-job/{out,err}.log"))
	})

	It("rejects custom path with shell metacharacters", func() {
		_, err := cmd.BuildErrandTailCmd("smoke-tests", "foo; rm -rf /")
		Expect(err).To(MatchError(ContainSubstring("invalid characters")))
	})

	It("rejects errand name with shell metacharacters", func() {
		_, err := cmd.BuildErrandTailCmd("bad$(cmd)", "")
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring("invalid characters")))
	})
})
