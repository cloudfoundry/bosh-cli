package cmd_test

import (
	"errors"
	"fmt"
	"os"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/clock/fakeclock"
	"github.com/cloudfoundry/bosh-agent/agentclient"
	mockhttpagent "github.com/cloudfoundry/bosh-agent/agentclient/http/mocks"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mockagentclient "github.com/cloudfoundry/bosh-cli/v7/agentclient/mocks"
	. "github.com/cloudfoundry/bosh-cli/v7/cmd"
	fakecmd "github.com/cloudfoundry/bosh-cli/v7/cmd/cmdfakes"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/mocks"
	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	boshssh "github.com/cloudfoundry/bosh-cli/v7/ssh"
	fakessh "github.com/cloudfoundry/bosh-cli/v7/ssh/sshfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("Logs", func() {
	const UUID = "8c5ff117-9572-45c5-8564-8bcf076ecafa"
	const ExpUsername = "bosh_8c5ff117957245c"

	Describe("LogsCmd", func() {
		var (
			deployment      *fakedir.FakeDeployment
			downloader      *fakecmd.FakeDownloader
			uuidGen         *fakeuuid.FakeGenerator
			nonIntSSHRunner *fakessh.FakeRunner
			command         LogsCmd
		)

		BeforeEach(func() {
			deployment = &fakedir.FakeDeployment{
				NameStub: func() string { return "dep" },
			}
			downloader = &fakecmd.FakeDownloader{}
			uuidGen = &fakeuuid.FakeGenerator{}
			nonIntSSHRunner = &fakessh.FakeRunner{}
			command = NewLogsCmd(deployment, downloader, uuidGen, nonIntSSHRunner)
		})

		Describe("Run", func() {
			var (
				opts LogsOpts
			)

			BeforeEach(func() {
				opts = LogsOpts{
					Args: AllOrInstanceGroupOrInstanceSlugArgs{
						Slug: boshdir.NewAllOrInstanceGroupOrInstanceSlug("job", "index"),
					},

					Directory: DirOrCWDArg{Path: "/fake-dir"},
				}
			})

			act := func() error { return command.Run(opts) }

			Context("when fetching logs (not tailing)", func() {
				It("fetches logs for a given instance", func() {
					result := boshdir.LogsResult{BlobstoreID: "blob-id", SHA1: "sha1"}
					deployment.FetchLogsReturns(result, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(deployment.FetchLogsCallCount()).To(Equal(1))

					slug, filters, agent, system, all := deployment.FetchLogsArgsForCall(0)
					Expect(slug).To(Equal(boshdir.NewAllOrInstanceGroupOrInstanceSlug("job", "index")))
					Expect(filters).To(BeEmpty())
					Expect(agent).To(BeFalse())
					Expect(system).To(BeFalse())
					Expect(all).To(BeFalse())

					Expect(downloader.DownloadCallCount()).To(Equal(1))

					blobID, sha1, prefix, dstDirPath := downloader.DownloadArgsForCall(0)
					Expect(blobID).To(Equal("blob-id"))
					Expect(sha1).To(Equal("sha1"))
					Expect(prefix).To(Equal("dep.job.index"))
					Expect(dstDirPath).To(Equal("/fake-dir"))
				})

				It("fetches agent logs and allows custom filters", func() {
					opts.Filters = []string{"filter1", "filter2"}
					opts.Agent = true

					deployment.FetchLogsReturns(boshdir.LogsResult{}, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(deployment.FetchLogsCallCount()).To(Equal(1))

					slug, filters, agent, system, all := deployment.FetchLogsArgsForCall(0)
					Expect(slug).To(Equal(boshdir.NewAllOrInstanceGroupOrInstanceSlug("job", "index")))
					Expect(filters).To(Equal([]string{"filter1", "filter2"}))
					Expect(agent).To(BeTrue())
					Expect(system).To(BeFalse())
					Expect(all).To(BeFalse())
				})

				It("fetches system logs and allows custom filters", func() {
					opts.Filters = []string{"filter1", "filter2"}
					opts.System = true

					deployment.FetchLogsReturns(boshdir.LogsResult{}, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(deployment.FetchLogsCallCount()).To(Equal(1))

					slug, filters, agent, system, all := deployment.FetchLogsArgsForCall(0)
					Expect(slug).To(Equal(boshdir.NewAllOrInstanceGroupOrInstanceSlug("job", "index")))
					Expect(filters).To(Equal([]string{"filter1", "filter2"}))
					Expect(agent).To(BeFalse())
					Expect(system).To(BeTrue())
					Expect(all).To(BeFalse())
				})

				It("fetches all logs and allows custom filters", func() {
					opts.Filters = []string{"filter1", "filter2"}
					opts.All = true

					deployment.FetchLogsReturns(boshdir.LogsResult{}, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(deployment.FetchLogsCallCount()).To(Equal(1))

					slug, filters, agent, system, all := deployment.FetchLogsArgsForCall(0)
					Expect(slug).To(Equal(boshdir.NewAllOrInstanceGroupOrInstanceSlug("job", "index")))
					Expect(filters).To(Equal([]string{"filter1", "filter2"}))
					Expect(agent).To(BeFalse())
					Expect(system).To(BeFalse())
					Expect(all).To(BeTrue())
				})

				It("fetches logs for more than one instance", func() {
					opts.Args.Slug = boshdir.NewAllOrInstanceGroupOrInstanceSlug("", "")

					result := boshdir.LogsResult{BlobstoreID: "blob-id", SHA1: "sha1"}
					deployment.FetchLogsReturns(result, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())

					Expect(deployment.FetchLogsCallCount()).To(Equal(1))

					Expect(downloader.DownloadCallCount()).To(Equal(1))

					blobID, sha1, prefix, dstDirPath := downloader.DownloadArgsForCall(0)
					Expect(blobID).To(Equal("blob-id"))
					Expect(sha1).To(Equal("sha1"))
					Expect(prefix).To(Equal("dep"))
					Expect(dstDirPath).To(Equal("/fake-dir"))
				})

				It("returns error if fetching logs failed", func() {
					deployment.FetchLogsReturns(boshdir.LogsResult{}, errors.New("fake-err"))

					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))
				})

				It("returns error if downloading release failed", func() {
					downloader.DownloadReturns(errors.New("fake-err"))

					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))
				})

				It("does not try to tail logs", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())
					Expect(nonIntSSHRunner.RunCallCount()).To(Equal(0))
				})
			})

			Context("when tailing logs (or specifying number of lines)", func() {

				BeforeEach(func() {
					opts.Follow = true
					opts.GatewayFlags.UUIDGen = uuidGen
					uuidGen.GeneratedUUID = UUID
				})

				It("sets up SSH access, runs SSH command and later cleans up SSH access", func() {
					nonIntSSHRunner.RunStub = func(boshssh.ConnectionOpts, boshdir.SSHResult, []string) error {
						Expect(deployment.CleanUpSSHCallCount()).To(Equal(0))
						return nil
					}
					Expect(act()).ToNot(HaveOccurred())

					Expect(deployment.SetUpSSHCallCount()).To(Equal(1))
					Expect(nonIntSSHRunner.RunCallCount()).To(Equal(1))
					Expect(deployment.CleanUpSSHCallCount()).To(Equal(1))

					setupSlug, setupSSHOpts := deployment.SetUpSSHArgsForCall(0)
					Expect(setupSlug).To(Equal(boshdir.NewAllOrInstanceGroupOrInstanceSlug("job", "index")))
					Expect(setupSSHOpts.Username).To(Equal(ExpUsername))
					Expect(setupSSHOpts.PublicKey).To(ContainSubstring("ssh-rsa AAAA"))

					slug, sshOpts := deployment.CleanUpSSHArgsForCall(0)
					Expect(slug).To(Equal(setupSlug))
					Expect(sshOpts).To(Equal(setupSSHOpts))
				})

				It("sets up SSH access for more than one instance", func() {
					opts.Args.Slug = boshdir.NewAllOrInstanceGroupOrInstanceSlug("", "")

					Expect(act()).ToNot(HaveOccurred())

					setupSlug, _ := deployment.SetUpSSHArgsForCall(0)
					Expect(setupSlug).To(Equal(boshdir.NewAllOrInstanceGroupOrInstanceSlug("", "")))
				})

				It("runs non-interactive SSH", func() {
					Expect(act()).ToNot(HaveOccurred())
					Expect(nonIntSSHRunner.RunCallCount()).To(Equal(1))
				})

				It("returns an error if generating SSH options fails", func() {
					uuidGen.GenerateError = errors.New("fake-err")
					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))
				})

				It("returns an error if setting up SSH access fails", func() {
					deployment.SetUpSSHReturns(boshdir.SSHResult{}, errors.New("fake-err"))
					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))
				})

				It("runs non-interactive SSH session with flags, and basic tail -f command that tails all logs", func() {
					result := boshdir.SSHResult{Hosts: []boshdir.Host{{Host: "ip1"}}}
					deployment.SetUpSSHReturns(result, nil)

					opts.GatewayFlags.Disable = true
					opts.GatewayFlags.Username = "gw-username"
					opts.GatewayFlags.Host = "gw-host"
					opts.GatewayFlags.PrivateKeyPath = "gw-private-key"
					opts.GatewayFlags.SOCKS5Proxy = "some-proxy"

					Expect(act()).ToNot(HaveOccurred())

					Expect(nonIntSSHRunner.RunCallCount()).To(Equal(1))

					runConnOpts, runResult, runCommand := nonIntSSHRunner.RunArgsForCall(0)
					Expect(runConnOpts.PrivateKey).To(ContainSubstring("-----BEGIN RSA PRIVATE KEY-----"))
					Expect(runConnOpts.GatewayDisable).To(Equal(true))
					Expect(runConnOpts.GatewayUsername).To(Equal("gw-username"))
					Expect(runConnOpts.GatewayHost).To(Equal("gw-host"))
					Expect(runConnOpts.GatewayPrivateKeyPath).To(Equal("gw-private-key"))
					Expect(runConnOpts.SOCKS5Proxy).To(Equal("some-proxy"))
					Expect(runResult).To(Equal(boshdir.SSHResult{Hosts: []boshdir.Host{{Host: "ip1"}}}))
					Expect(runCommand).To(Equal([]string{"sudo", "bash", "-c", "'exec tail -F /var/vcap/sys/log/**/*.log $(if [ -f /var/vcap/sys/log/*.log ]; then echo /var/vcap/sys/log/*.log ; fi)'"}))
				})

				It("runs tail command with specified number of lines and quiet option", func() {
					opts.Num = 10
					opts.Quiet = true

					deployment.SetUpSSHReturns(boshdir.SSHResult{}, nil)
					Expect(act()).ToNot(HaveOccurred())

					_, _, runCommand := nonIntSSHRunner.RunArgsForCall(0)
					Expect(runCommand).To(Equal([]string{
						"sudo", "bash", "-c", "'exec tail -F -n 10 -q /var/vcap/sys/log/**/*.log $(if [ -f /var/vcap/sys/log/*.log ]; then echo /var/vcap/sys/log/*.log ; fi)'"}))
				})

				It("runs tail command with specified number of lines even if following is not requested", func() {
					opts.Follow = false
					opts.Num = 10

					deployment.SetUpSSHReturns(boshdir.SSHResult{}, nil)
					Expect(act()).ToNot(HaveOccurred())

					_, _, runCommand := nonIntSSHRunner.RunArgsForCall(0)
					Expect(runCommand).To(Equal([]string{
						"sudo", "bash", "-c", "'exec tail -n 10 /var/vcap/sys/log/**/*.log $(if [ -f /var/vcap/sys/log/*.log ]; then echo /var/vcap/sys/log/*.log ; fi)'"}))
				})

				It("runs tail command for the agent log if agent is specified", func() {
					opts.Agent = true

					deployment.SetUpSSHReturns(boshdir.SSHResult{}, nil)
					Expect(act()).ToNot(HaveOccurred())

					_, _, runCommand := nonIntSSHRunner.RunArgsForCall(0)
					Expect(runCommand).To(Equal([]string{
						"sudo", "bash", "-c", "'exec tail -F /var/vcap/bosh/log/current'"}))
				})

				It("runs tail command with jobs filters if specified", func() {
					opts.Jobs = []string{"job1", "job2"}

					deployment.SetUpSSHReturns(boshdir.SSHResult{}, nil)
					Expect(act()).ToNot(HaveOccurred())

					_, _, runCommand := nonIntSSHRunner.RunArgsForCall(0)
					Expect(runCommand).To(Equal([]string{
						"sudo", "bash", "-c", "'exec tail -F /var/vcap/sys/log/job1/*.log /var/vcap/sys/log/job2/*.log'"}))
				})

				It("runs tail command with custom filters if specified", func() {
					opts.Filters = []string{"other/*.log", "**/*.log"}

					deployment.SetUpSSHReturns(boshdir.SSHResult{}, nil)
					Expect(act()).ToNot(HaveOccurred())

					_, _, runCommand := nonIntSSHRunner.RunArgsForCall(0)
					Expect(runCommand).To(Equal([]string{
						"sudo", "bash", "-c", "'exec tail -F /var/vcap/sys/log/other/*.log /var/vcap/sys/log/**/*.log'"}))
				})

				It("runs tail command with agent log, and custom filters", func() {
					opts.Agent = true
					opts.Filters = []string{"other/*.log", "**/*.log"}

					deployment.SetUpSSHReturns(boshdir.SSHResult{}, nil)
					Expect(act()).ToNot(HaveOccurred())

					_, _, runCommand := nonIntSSHRunner.RunArgsForCall(0)
					Expect(runCommand).To(Equal([]string{
						"sudo", "bash", "-c", "'exec tail -F /var/vcap/bosh/log/current /var/vcap/sys/log/other/*.log /var/vcap/sys/log/**/*.log'"}))
				})

				It("returns error if non-interactive SSH session errors", func() {
					nonIntSSHRunner.RunReturns(errors.New("fake-err"))

					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))
				})

				It("does not try to fetch logs", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())
					Expect(deployment.FetchLogsCallCount()).To(Equal(0))
				})
			})
		})
	})

	Describe("EnvLogsCmd", func() {
		var (
			mockCtrl *gomock.Controller

			agentClientFactory *mockhttpagent.MockAgentClientFactory
			agentClient        *mockagentclient.MockAgentClient
			nonIntSSHRunner    *fakessh.FakeRunner
			scpRunner          *fakessh.FakeSCPRunner
			fs                 *fakes.FakeFileSystem
			timeService        clock.Clock
			ui                 *fakeui.FakeUI

			uuidGen *fakeuuid.FakeGenerator

			command EnvLogsCmd
		)

		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())

			agentClient = mockagentclient.NewMockAgentClient(mockCtrl)
			agentClientFactory = mockhttpagent.NewMockAgentClientFactory(mockCtrl)
			nonIntSSHRunner = &fakessh.FakeRunner{}
			scpRunner = &fakessh.FakeSCPRunner{}
			fs = fakes.NewFakeFileSystem()
			timeService = fakeclock.NewFakeClock(time.Date(2009, time.November, 10, 23, 1, 2, 333, time.UTC))
			ui = &fakeui.FakeUI{}

			uuidGen = &fakeuuid.FakeGenerator{}

			command = NewEnvLogsCmd(agentClientFactory, nonIntSSHRunner, scpRunner, fs, timeService, ui)
		})

		AfterEach(func() {
			mockCtrl.Finish()
		})

		Describe("Run", func() {
			var (
				opts LogsOpts
			)

			Context("neither the endpoint or certificate flag is set", func() {
				BeforeEach(func() {
					opts = LogsOpts{
						CreateEnvAuthFlags: CreateEnvAuthFlags{
							TargetDirector: true,
						},
					}
				})

				It("errors", func() {
					Expect(command.Run(opts)).To(MatchError("the --director flag requires both the --agent-endpoint and --agent-certificate flags to be set"))
				})
			})

			Context("only the endpoint flag is set", func() {
				BeforeEach(func() {
					opts = LogsOpts{
						CreateEnvAuthFlags: CreateEnvAuthFlags{
							TargetDirector: true,
							Endpoint:       "https:///foo:bar@10.0.0.5",
						},
					}
				})

				It("errors", func() {
					Expect(command.Run(opts)).To(MatchError("the --director flag requires both the --agent-endpoint and --agent-certificate flags to be set"))
				})
			})

			Context("only the certificate flag is set", func() {
				BeforeEach(func() {
					opts = LogsOpts{
						CreateEnvAuthFlags: CreateEnvAuthFlags{
							TargetDirector: true,
							Certificate:    "some-cert",
						},
					}
				})

				It("errors", func() {
					Expect(command.Run(opts)).To(MatchError("the --director flag requires both the --agent-endpoint and --agent-certificate flags to be set"))
				})
			})

			Context("the endpoint and certificate flags are set", func() {
				BeforeEach(func() {
					opts = LogsOpts{
						CreateEnvAuthFlags: CreateEnvAuthFlags{
							TargetDirector: true,
							Endpoint:       "https:///foo:bar@10.0.0.5",
							Certificate:    "some-cert",
						},
						GatewayFlags: GatewayFlags{UUIDGen: uuidGen},
					}

					uuidGen.GeneratedUUID = UUID

					agentClientFactory.EXPECT().NewAgentClient(
						gomock.Eq("bosh-cli"),
						gomock.Eq("https:///foo:bar@10.0.0.5"),
						gomock.Eq("some-cert"),
					).Return(agentClient, nil).Times(1)
				})

				It("returns an error if generating SSH options fails", func() {
					uuidGen.GenerateError = errors.New("fake-err")

					err := command.Run(opts)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))
				})

				It("returns an error if setting up SSH access fails", func() {
					agentClient.EXPECT().SetUpSSH(gomock.Any(), gomock.Any()).Return(agentclient.SSHResult{}, errors.New("fake-ssh-err"))

					err := command.Run(opts)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-ssh-err"))
				})

				Context("tailing logs", func() {
					BeforeEach(func() {
						opts.Follow = true
					})

					It("sets up SSH access, runs SSH command and later cleans up SSH access", func() {
						nonIntSSHRunner.RunStub = func(boshssh.ConnectionOpts, boshdir.SSHResult, []string) error {
							agentClient.EXPECT().CleanUpSSH(gomock.Any()).Times(0)
							return nil
						}

						agentClient.EXPECT().SetUpSSH(gomock.Eq(ExpUsername), mocks.GomegaMock(ContainSubstring("ssh-rsa AAAA"))).
							Times(1)
						agentClient.EXPECT().CleanUpSSH(gomock.Eq(ExpUsername)).
							Times(1)

						Expect(command.Run(opts)).ToNot(HaveOccurred())

						Expect(nonIntSSHRunner.RunCallCount()).To(Equal(1))
					})

					It("runs non-interactive SSH session with flags, and basic tail -f command that tails all logs", func() {
						result := agentclient.SSHResult{
							Command:       "setup",
							Status:        "success",
							Ip:            "10.0.0.5",
							HostPublicKey: "some-public-key",
						}
						agentClient.EXPECT().SetUpSSH(gomock.Any(), gomock.Any()).Return(result, nil)
						agentClient.EXPECT().CleanUpSSH(gomock.Any()).Times(1)

						opts.GatewayFlags.Disable = true
						opts.GatewayFlags.Username = "gw-username"
						opts.GatewayFlags.Host = "gw-host"
						opts.GatewayFlags.PrivateKeyPath = "gw-private-key"
						opts.GatewayFlags.SOCKS5Proxy = "some-proxy"

						Expect(command.Run(opts)).ToNot(HaveOccurred())

						Expect(nonIntSSHRunner.RunCallCount()).To(Equal(1))

						runConnOpts, runResult, runCommand := nonIntSSHRunner.RunArgsForCall(0)
						Expect(runConnOpts.PrivateKey).To(ContainSubstring("-----BEGIN RSA PRIVATE KEY-----"))
						Expect(runConnOpts.GatewayDisable).To(Equal(true))
						Expect(runConnOpts.GatewayUsername).To(Equal("gw-username"))
						Expect(runConnOpts.GatewayHost).To(Equal("gw-host"))
						Expect(runConnOpts.GatewayPrivateKeyPath).To(Equal("gw-private-key"))
						Expect(runConnOpts.SOCKS5Proxy).To(Equal("some-proxy"))
						Expect(runResult).To(Equal(boshdir.SSHResult{Hosts: []boshdir.Host{{Username: ExpUsername, Host: "10.0.0.5", HostPublicKey: "some-public-key", Job: "create-env-vm", IndexOrID: "0"}}}))
						Expect(runCommand).To(Equal([]string{"sudo", "bash", "-c", "'exec tail -F /var/vcap/sys/log/**/*.log $(if [ -f /var/vcap/sys/log/*.log ]; then echo /var/vcap/sys/log/*.log ; fi)'"}))
					})

					Context("tail options", func() {
						BeforeEach(func() {
							agentClient.EXPECT().SetUpSSH(gomock.Any(), gomock.Any()).Return(agentclient.SSHResult{}, nil).Times(1)
							agentClient.EXPECT().CleanUpSSH(gomock.Any()).Times(1)
						})

						It("runs tail command with specified number of lines and quiet option", func() {
							opts.Num = 10
							opts.Quiet = true

							Expect(command.Run(opts)).ToNot(HaveOccurred())

							_, _, runCommand := nonIntSSHRunner.RunArgsForCall(0)
							Expect(runCommand).To(Equal([]string{
								"sudo", "bash", "-c", "'exec tail -F -n 10 -q /var/vcap/sys/log/**/*.log $(if [ -f /var/vcap/sys/log/*.log ]; then echo /var/vcap/sys/log/*.log ; fi)'"}))
						})

						It("runs tail command with specified number of lines even if following is not requested", func() {
							opts.Follow = false
							opts.Num = 10

							Expect(command.Run(opts)).ToNot(HaveOccurred())

							_, _, runCommand := nonIntSSHRunner.RunArgsForCall(0)
							Expect(runCommand).To(Equal([]string{
								"sudo", "bash", "-c", "'exec tail -n 10 /var/vcap/sys/log/**/*.log $(if [ -f /var/vcap/sys/log/*.log ]; then echo /var/vcap/sys/log/*.log ; fi)'"}))
						})

						It("runs tail command for the agent log if agent is specified", func() {
							opts.Agent = true

							Expect(command.Run(opts)).ToNot(HaveOccurred())

							_, _, runCommand := nonIntSSHRunner.RunArgsForCall(0)
							Expect(runCommand).To(Equal([]string{
								"sudo", "bash", "-c", "'exec tail -F /var/vcap/bosh/log/current'"}))
						})

						It("runs tail command with jobs filters if specified", func() {
							opts.Jobs = []string{"job1", "job2"}

							Expect(command.Run(opts)).ToNot(HaveOccurred())

							_, _, runCommand := nonIntSSHRunner.RunArgsForCall(0)
							Expect(runCommand).To(Equal([]string{
								"sudo", "bash", "-c", "'exec tail -F /var/vcap/sys/log/job1/*.log /var/vcap/sys/log/job2/*.log'"}))
						})

						It("runs tail command with custom filters if specified", func() {
							opts.Filters = []string{"other/*.log", "**/*.log"}

							Expect(command.Run(opts)).ToNot(HaveOccurred())

							_, _, runCommand := nonIntSSHRunner.RunArgsForCall(0)
							Expect(runCommand).To(Equal([]string{
								"sudo", "bash", "-c", "'exec tail -F /var/vcap/sys/log/other/*.log /var/vcap/sys/log/**/*.log'"}))
						})

						It("runs tail command with agent log, and custom filters", func() {
							opts.Agent = true
							opts.Filters = []string{"other/*.log", "**/*.log"}

							Expect(command.Run(opts)).ToNot(HaveOccurred())

							_, _, runCommand := nonIntSSHRunner.RunArgsForCall(0)
							Expect(runCommand).To(Equal([]string{
								"sudo", "bash", "-c", "'exec tail -F /var/vcap/bosh/log/current /var/vcap/sys/log/other/*.log /var/vcap/sys/log/**/*.log'"}))
						})

						It("returns error if non-interactive SSH session errors", func() {
							nonIntSSHRunner.RunReturns(errors.New("fake-err"))

							err := command.Run(opts)
							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("fake-err"))
						})

						It("does not try to fetch logs", func() {
							agentClient.EXPECT().BundleLogs(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

							Expect(command.Run(opts)).ToNot(HaveOccurred())
						})
					})
				})

				Context("fetching logs", func() {
					const emptyFileSHA512 string = "sha512:cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"
					var (
						bundleResult agentclient.BundleLogsResult
					)

					BeforeEach(func() {
						agentClient.EXPECT().SetUpSSH(gomock.Any(), gomock.Any()).Return(agentclient.SSHResult{}, nil).Times(1)
						agentClient.EXPECT().CleanUpSSH(gomock.Any()).Times(1)

						bundleResult = agentclient.BundleLogsResult{
							LogsTarPath:  "/foo/bar",
							SHA512Digest: emptyFileSHA512,
						}
					})

					It("bundles logs for jobs by default", func() {
						agentClient.EXPECT().BundleLogs(
							gomock.Eq(ExpUsername),
							gomock.Eq("job"),
							mocks.GomegaMock(HaveLen(0)),
						).
							Return(bundleResult, nil).
							Times(1)
						agentClient.EXPECT().RemoveFile(gomock.Any()).
							Times(1)

						Expect(command.Run(opts)).ToNot(HaveOccurred())
					})

					It("bundles agent logs and allows custom filters", func() {
						opts.Filters = []string{"filter1", "filter2"}
						opts.Agent = true

						agentClient.EXPECT().BundleLogs(
							gomock.Eq(ExpUsername),
							gomock.Eq("agent"),
							gomock.Eq([]string{"filter1", "filter2"}),
						).
							Return(bundleResult, nil).
							Times(1)
						agentClient.EXPECT().RemoveFile(gomock.Any()).
							Times(1)

						Expect(command.Run(opts)).ToNot(HaveOccurred())
					})

					It("returns error if bundling logs failed", func() {
						agentClient.EXPECT().BundleLogs(gomock.Any(), gomock.Any(), gomock.Any()).Return(agentclient.BundleLogsResult{}, errors.New("fake-logs-err"))

						err := command.Run(opts)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-logs-err"))
					})

					It("uses scp to download the log bundle", func() {
						fakeFile := fakes.NewFakeFile("/tmp/baz", fs)
						fs.ReturnTempFile = fakeFile
						agentClient.EXPECT().BundleLogs(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(bundleResult, nil).
							Times(1)
						agentClient.EXPECT().RemoveFile(gomock.Any()).
							Times(1)

						Expect(command.Run(opts)).ToNot(HaveOccurred())

						Expect(scpRunner.RunCallCount()).To(Equal(1))
						_, _, scpArgs := scpRunner.RunArgsForCall(0)
						slug, err := scpArgs.AllOrInstanceGroupOrInstanceSlug()
						Expect(err).NotTo(HaveOccurred())
						Expect(slug.String()).To(Equal("create-env-vm/0"))
						scpHost := boshdir.Host{Username: ExpUsername, Host: "10.0.0.5", Job: "create-env-vm", IndexOrID: "0"}
						Expect(scpArgs.ForHost(scpHost)).To(HaveExactElements(fmt.Sprintf("%s@10.0.0.5:/foo/bar", ExpUsername), "/tmp/baz"))

						Expect(ui.Said).To(ContainElement("Downloading create-env-vm/0 logs to 'create-env-vm-logs-20091110-230102-000000333.tgz'..."))
					})

					It("returns error if scp fails", func() {
						scpRunner.RunReturns(errors.New("fake-scp-error"))
						agentClient.EXPECT().BundleLogs(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(bundleResult, nil).
							Times(1)
						agentClient.EXPECT().RemoveFile(gomock.Any()).
							Times(1)

						err := command.Run(opts)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Running SCP"))
						Expect(err.Error()).To(ContainSubstring("fake-scp-err"))
					})

					It("returns error if parsing the sha fails", func() {
						bundleResult.SHA512Digest = "garbage can"

						agentClient.EXPECT().BundleLogs(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(bundleResult, nil).
							Times(1)
						agentClient.EXPECT().RemoveFile(gomock.Any()).
							Times(1)

						err := command.Run(opts)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Unable to parse digest string"))
					})

					It("returns error if verifying the log file sha fails", func() {
						fakeFile := fakes.NewFakeFile("/tmp/baz", fs)
						fakeFile.Write([]byte("not empty anymore!")) //nolint:errcheck
						fs.ReturnTempFile = fakeFile

						agentClient.EXPECT().BundleLogs(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(bundleResult, nil).
							Times(1)
						agentClient.EXPECT().RemoveFile(gomock.Any()).
							Times(1)

						err := command.Run(opts)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Expected stream to have digest"))
					})

					It("Respects the path arg", func() {
						agentClient.EXPECT().BundleLogs(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(bundleResult, nil).
							Times(1)
						agentClient.EXPECT().RemoveFile(gomock.Any()).
							Times(1)
						opts.Directory.Path = "/hey/hello"
						fs.MkdirAll("/hey/hello", os.FileMode(0777)) //nolint:errcheck
						fs.ReturnTempFilesByPrefix = map[string]boshsys.File{
							"bosh-cli-scp-download": fakes.NewFakeFile("/tmp/baz", fs),
						}
						fs.RenameStub = func(oldPath, newPath string) error {
							Expect(oldPath).To(Equal("/tmp/baz"))
							Expect(newPath).To(Equal("/hey/hello/create-env-vm-logs-20091110-230102-000000333.tgz"))

							return nil
						}

						Expect(command.Run(opts)).ToNot(HaveOccurred())
						Expect(ui.Said).To(ContainElement("Downloading create-env-vm/0 logs to '/hey/hello/create-env-vm-logs-20091110-230102-000000333.tgz'..."))
					})

					It("returns error if closing temp file fails", func() {
						fakeFile := fakes.NewFakeFile("/tmp/baz", fs)
						fakeFile.CloseErr = errors.New("fake-close-error")
						fs.ReturnTempFile = fakeFile
						agentClient.EXPECT().BundleLogs(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(bundleResult, nil).
							Times(1)
						agentClient.EXPECT().RemoveFile(gomock.Any()).
							Times(1)

						err := command.Run(opts)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-close-err"))
					})

					It("returns error if moving file fails", func() {
						fs.RenameError = errors.New("fake-rename-error")
						agentClient.EXPECT().BundleLogs(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(bundleResult, nil).
							Times(1)
						agentClient.EXPECT().RemoveFile(gomock.Any()).
							Times(1)

						err := command.Run(opts)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Moving to final destination"))
						Expect(err.Error()).To(ContainSubstring("fake-rename-err"))
					})

					It("does not try to tail logs", func() {
						agentClient.EXPECT().BundleLogs(gomock.Any(), gomock.Any(), gomock.Any()).
							Return(bundleResult, nil).
							Times(1)
						agentClient.EXPECT().RemoveFile(gomock.Any()).
							Times(1)

						Expect(command.Run(opts)).ToNot(HaveOccurred())
						Expect(nonIntSSHRunner.RunCallCount()).To(Equal(0))
					})
				})
			})
		})
	})
})
