package cmd_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-agent/v2/agentclient"
	mockhttpagent "github.com/cloudfoundry/bosh-agent/v2/agentclient/http/mocks"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mockagentclient "github.com/cloudfoundry/bosh-cli/v7/agentclient/mocks"
	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/mocks"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	boshssh "github.com/cloudfoundry/bosh-cli/v7/ssh"
	fakessh "github.com/cloudfoundry/bosh-cli/v7/ssh/sshfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("SSH", func() {
	const UUID = "8c5ff117-9572-45c5-8564-8bcf076ecafa"
	const ExpUsername = "bosh_8c5ff117957245c"

	Describe("SSHCmd", func() {
		var (
			deployment       *fakedir.FakeDeployment
			uuidGen          *fakeuuid.FakeGenerator
			intSSHRunner     *fakessh.FakeRunner
			nonIntSSHRunner  *fakessh.FakeRunner
			resultsSSHRunner *fakessh.FakeRunner
			ui               *fakeui.FakeUI
			hostBuilder      *fakessh.FakeHostBuilder
			command          cmd.SSHCmd
		)

		BeforeEach(func() {
			deployment = &fakedir.FakeDeployment{}
			uuidGen = &fakeuuid.FakeGenerator{}
			intSSHRunner = &fakessh.FakeRunner{}
			nonIntSSHRunner = &fakessh.FakeRunner{}
			resultsSSHRunner = &fakessh.FakeRunner{}
			hostBuilder = &fakessh.FakeHostBuilder{}
			ui = &fakeui.FakeUI{}
			command = cmd.NewSSHCmd(intSSHRunner, nonIntSSHRunner, resultsSSHRunner, ui, hostBuilder)
		})

		Describe("Run", func() {
			var (
				sshOpts opts.SSHOpts
				act     func() error
			)

			BeforeEach(func() {
				sshOpts = opts.SSHOpts{
					Args: opts.SshSlugArgs{
						Slug: boshdir.NewAllOrInstanceGroupOrInstanceSlug("job-name", ""),
					},

					GatewayFlags: opts.GatewayFlags{
						UUIDGen: uuidGen,
					},
				}

				uuidGen.GeneratedUUID = UUID

				act = func() error {
					return command.Run(sshOpts, func() (boshdir.Deployment, error) {
						return deployment, nil
					})
				}
			})

			itRunsNonInteractiveSSHWhenCommandIsGiven := func(runner **fakessh.FakeRunner) {
				Context("when command is provided", func() {
					BeforeEach(func() {
						sshOpts.Command = []string{"cmd", "arg1"}
					})

					It("sets up SSH access, runs SSH command and later cleans up SSH access", func() {
						(*runner).RunStub = func(boshssh.ConnectionOpts, boshdir.SSHResult, []string) error {
							Expect(deployment.CleanUpSSHCallCount()).To(Equal(0))
							return nil
						}
						Expect(act()).ToNot(HaveOccurred())

						Expect(deployment.SetUpSSHCallCount()).To(Equal(1))
						Expect((*runner).RunCallCount()).To(Equal(1))
						Expect(deployment.CleanUpSSHCallCount()).To(Equal(1))

						setupSlug, setupSSHOpts := deployment.SetUpSSHArgsForCall(0)
						Expect(setupSlug).To(Equal(boshdir.NewAllOrInstanceGroupOrInstanceSlug("job-name", "")))
						Expect(setupSSHOpts.Username).To(Equal(ExpUsername))
						Expect(setupSSHOpts.PublicKey).To(ContainSubstring("ssh-rsa AAAA"))

						slug, sshOpts := deployment.CleanUpSSHArgsForCall(0)
						Expect(slug).To(Equal(setupSlug))
						Expect(sshOpts).To(Equal(setupSSHOpts))
					})

					It("runs non-interactive SSH", func() {
						Expect(act()).ToNot(HaveOccurred())
						Expect((*runner).RunCallCount()).To(Equal(1))
						Expect(intSSHRunner.RunCallCount()).To(Equal(0))
					})

					It("returns an error if setting up SSH access fails", func() {
						deployment.SetUpSSHReturns(boshdir.SSHResult{}, errors.New("fake-err"))
						err := act()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-err"))
					})

					It("returns an error if generating SSH options fails", func() {
						uuidGen.GenerateError = errors.New("fake-err")
						err := act()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-err"))
					})

					It("runs non-interactive SSH session with flags, and command", func() {
						result := boshdir.SSHResult{Hosts: []boshdir.Host{{Host: "ip1"}}}
						deployment.SetUpSSHReturns(result, nil)

						sshOpts.RawOpts = opts.TrimmedSpaceArgs([]string{"raw1", "raw2"})
						sshOpts.GatewayFlags.Disable = true
						sshOpts.GatewayFlags.Username = "gw-username"
						sshOpts.GatewayFlags.Host = "gw-host"
						sshOpts.GatewayFlags.PrivateKeyPath = "gw-private-key"
						sshOpts.GatewayFlags.SOCKS5Proxy = "socks5"

						Expect(act()).ToNot(HaveOccurred())

						Expect((*runner).RunCallCount()).To(Equal(1))

						runConnOpts, runResult, runCommand := (*runner).RunArgsForCall(0)
						Expect(runConnOpts.RawOpts).To(Equal([]string{"raw1", "raw2", "-o", "StrictHostKeyChecking=yes"}))
						Expect(runConnOpts.PrivateKey).To(ContainSubstring("-----BEGIN RSA PRIVATE KEY-----"))
						Expect(runConnOpts.GatewayDisable).To(Equal(true))
						Expect(runConnOpts.GatewayUsername).To(Equal("gw-username"))
						Expect(runConnOpts.GatewayHost).To(Equal("gw-host"))
						Expect(runConnOpts.GatewayPrivateKeyPath).To(Equal("gw-private-key"))
						Expect(runConnOpts.SOCKS5Proxy).To(Equal("socks5"))
						Expect(runResult).To(Equal(boshdir.SSHResult{Hosts: []boshdir.Host{{Host: "ip1"}}}))
						Expect(runCommand).To(Equal([]string{"cmd", "arg1"}))
					})

					It("returns error if non-interactive SSH session errors", func() {
						(*runner).RunReturns(errors.New("fake-err"))
						err := act()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-err"))
					})
				})
			}

			Context("when ui is interactive", func() {
				BeforeEach(func() {
					ui.Interactive = true
				})

				itRunsNonInteractiveSSHWhenCommandIsGiven(&nonIntSSHRunner)

				Context("when command is not provided", func() {
					It("sets up SSH access, runs SSH command and later cleans up SSH access", func() {
						intSSHRunner.RunStub = func(boshssh.ConnectionOpts, boshdir.SSHResult, []string) error {
							Expect(deployment.CleanUpSSHCallCount()).To(Equal(0))
							return nil
						}
						Expect(act()).ToNot(HaveOccurred())

						Expect(deployment.SetUpSSHCallCount()).To(Equal(1))
						Expect(intSSHRunner.RunCallCount()).To(Equal(1))
						Expect(deployment.CleanUpSSHCallCount()).To(Equal(1))

						setupSlug, setupSSHOpts := deployment.SetUpSSHArgsForCall(0)
						Expect(setupSlug).To(Equal(boshdir.NewAllOrInstanceGroupOrInstanceSlug("job-name", "")))
						Expect(setupSSHOpts.Username).To(Equal(ExpUsername))
						Expect(setupSSHOpts.PublicKey).To(ContainSubstring("ssh-rsa AAAA"))

						slug, sshOpts := deployment.CleanUpSSHArgsForCall(0)
						Expect(slug).To(Equal(setupSlug))
						Expect(sshOpts).To(Equal(setupSSHOpts))
					})

					It("runs only interactive SSH", func() {
						Expect(act()).ToNot(HaveOccurred())
						Expect(nonIntSSHRunner.RunCallCount()).To(Equal(0))
						Expect(intSSHRunner.RunCallCount()).To(Equal(1))
						Expect(resultsSSHRunner.RunCallCount()).To(Equal(0))
					})

					It("returns an error if setting up SSH access fails", func() {
						deployment.SetUpSSHReturns(boshdir.SSHResult{}, errors.New("fake-err"))
						err := act()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-err"))
					})

					It("runs interactive SSH session with flags, but without command", func() {
						result := boshdir.SSHResult{Hosts: []boshdir.Host{{Host: "ip1"}}}
						deployment.SetUpSSHReturns(result, nil)

						sshOpts.RawOpts = opts.TrimmedSpaceArgs([]string{"raw1", "raw2"})
						sshOpts.GatewayFlags.Disable = true
						sshOpts.GatewayFlags.Username = "gw-username"
						sshOpts.GatewayFlags.Host = "gw-host"
						sshOpts.GatewayFlags.PrivateKeyPath = "gw-private-key"

						Expect(act()).ToNot(HaveOccurred())

						Expect(intSSHRunner.RunCallCount()).To(Equal(1))

						runConnOpts, runResult, runCommand := intSSHRunner.RunArgsForCall(0)
						Expect(runConnOpts.RawOpts).To(Equal([]string{"raw1", "raw2", "-o", "StrictHostKeyChecking=yes"}))
						Expect(runConnOpts.PrivateKey).To(ContainSubstring("-----BEGIN RSA PRIVATE KEY-----"))
						Expect(runConnOpts.GatewayDisable).To(Equal(true))
						Expect(runConnOpts.GatewayUsername).To(Equal("gw-username"))
						Expect(runConnOpts.GatewayHost).To(Equal("gw-host"))
						Expect(runConnOpts.GatewayPrivateKeyPath).To(Equal("gw-private-key"))
						Expect(runResult).To(Equal(boshdir.SSHResult{Hosts: []boshdir.Host{{Host: "ip1"}}}))
						Expect(runCommand).To(BeNil())
					})

					It("returns error if interactive SSH session errors", func() {
						intSSHRunner.RunReturns(errors.New("fake-err"))
						err := act()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-err"))
					})
				})
			})

			Context("when ui is not interactive", func() {
				BeforeEach(func() {
					ui.Interactive = false
				})

				itRunsNonInteractiveSSHWhenCommandIsGiven(&nonIntSSHRunner)

				Context("when command is not provided", func() {
					It("returns an error since command is required", func() {
						Expect(act()).To(Equal(errors.New("Non-interactive SSH requires non-empty command")))
					})

					It("does not try to run any SSH sessions", func() {
						Expect(act()).To(HaveOccurred())
						Expect(intSSHRunner.RunCallCount()).To(Equal(0))
						Expect(nonIntSSHRunner.RunCallCount()).To(Equal(0))
						Expect(resultsSSHRunner.RunCallCount()).To(Equal(0))
					})
				})
			})

			Context("when results are requested", func() {
				BeforeEach(func() {
					ui.Interactive = true
					sshOpts.Results = true
				})

				itRunsNonInteractiveSSHWhenCommandIsGiven(&resultsSSHRunner)

				Context("when command is not provided", func() {
					It("returns an error since command is required", func() {
						Expect(act()).To(Equal(errors.New("Non-interactive SSH requires non-empty command")))
					})

					It("does not try to run any SSH sessions", func() {
						Expect(act()).To(HaveOccurred())
						Expect(intSSHRunner.RunCallCount()).To(Equal(0))
						Expect(nonIntSSHRunner.RunCallCount()).To(Equal(0))
						Expect(resultsSSHRunner.RunCallCount()).To(Equal(0))
					})
				})
			})

			Context("when private key is provided", func() {
				var expectedHost = boshdir.Host{
					Job:       "",
					IndexOrID: "",
					Username:  "vcap",
					Host:      "1.2.3.4",
				}
				BeforeEach(func() {
					ui.Interactive = false
					sshOpts.Command = []string{"do", "it"}

					sshOpts.PrivateKey.Bytes = []byte("topsecret")
					sshOpts.Username = "vcap"
					sshOpts.Args.Slug, _ = boshdir.NewAllOrInstanceGroupOrInstanceSlugFromString("1.2.3.4")

					hostBuilder.BuildHostReturns(expectedHost, nil)
				})

				It("agent is not used to setup ssh", func() {
					Expect(act()).ToNot(HaveOccurred())
					Expect(deployment.SetUpSSHCallCount()).To(Equal(0))
					Expect(deployment.CleanUpSSHCallCount()).To(Equal(0))
				})

				It("builds host from args", func() {
					Expect(act()).ToNot(HaveOccurred())

					Expect(hostBuilder.BuildHostCallCount()).To(Equal(1))
					slug, username, _ := hostBuilder.BuildHostArgsForCall(0)

					Expect(slug).To(Equal(sshOpts.Args.Slug))
					Expect(username).To(Equal(sshOpts.Username))

					Expect(nonIntSSHRunner.RunCallCount()).To(Equal(1))
					conn, result, _ := nonIntSSHRunner.RunArgsForCall(0)

					expectedResult := boshdir.SSHResult{
						Hosts: []boshdir.Host{
							expectedHost,
						},
						GatewayUsername: "",
						GatewayHost:     "",
					}
					Expect(result).To(Equal(expectedResult))

					expectedConn := boshssh.ConnectionOpts{
						PrivateKey: "topsecret",

						GatewayDisable: false,

						GatewayUsername:       "",
						GatewayHost:           "",
						GatewayPrivateKeyPath: "",

						SOCKS5Proxy: "",
						RawOpts:     []string{"-o", "StrictHostKeyChecking=no"},
					}
					Expect(conn).To(Equal(expectedConn))

				})
			})
		})
	})

	Describe("EnvSSHCmd", func() {

		var (
			mockCtrl *gomock.Controller

			agentClientFactory *mockhttpagent.MockAgentClientFactory
			agentClient        *mockagentclient.MockAgentClient
			intSSHRunner       *fakessh.FakeRunner
			nonIntSSHRunner    *fakessh.FakeRunner
			resultsSSHRunner   *fakessh.FakeRunner
			ui                 *fakeui.FakeUI

			uuidGen *fakeuuid.FakeGenerator

			command cmd.EnvSSHCmd
		)

		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())

			agentClient = mockagentclient.NewMockAgentClient(mockCtrl)
			agentClientFactory = mockhttpagent.NewMockAgentClientFactory(mockCtrl)
			intSSHRunner = &fakessh.FakeRunner{}
			nonIntSSHRunner = &fakessh.FakeRunner{}
			resultsSSHRunner = &fakessh.FakeRunner{}
			ui = &fakeui.FakeUI{}

			uuidGen = &fakeuuid.FakeGenerator{}

			command = cmd.NewEnvSSHCmd(agentClientFactory, intSSHRunner, nonIntSSHRunner, resultsSSHRunner, ui)
		})

		AfterEach(func() {
			mockCtrl.Finish()
		})

		Describe("Run", func() {
			var (
				sshOpts opts.SSHOpts
			)

			Context("when private key is provided", func() {
				BeforeEach(func() {
					sshOpts.PrivateKey.Bytes = []byte("topsecret")
				})

				It("errors", func() {
					err := command.Run(sshOpts)

					Expect(err).To(MatchError("the --private-key flag is not supported in combination with the --director flag"))
				})
			})

			Context("when private key is not provided", func() {
				Context("neither the endpoint or certificate flag is set", func() {
					BeforeEach(func() {
						sshOpts = opts.SSHOpts{
							CreateEnvAuthFlags: opts.CreateEnvAuthFlags{
								TargetDirector: true,
							},
						}
					})

					It("errors", func() {
						Expect(command.Run(sshOpts)).To(MatchError("the --director flag requires both the --agent-endpoint and --agent-certificate flags to be set"))
					})
				})

				Context("only the endpoint flag is set", func() {
					BeforeEach(func() {
						sshOpts = opts.SSHOpts{
							CreateEnvAuthFlags: opts.CreateEnvAuthFlags{
								TargetDirector: true,
								Endpoint:       "https:///foo:bar@10.0.0.5",
							},
						}
					})

					It("errors", func() {
						Expect(command.Run(sshOpts)).To(MatchError("the --director flag requires both the --agent-endpoint and --agent-certificate flags to be set"))
					})
				})

				Context("only the certificate flag is set", func() {
					BeforeEach(func() {
						sshOpts = opts.SSHOpts{
							CreateEnvAuthFlags: opts.CreateEnvAuthFlags{
								TargetDirector: true,
								Certificate:    "some-cert",
							},
						}
					})

					It("errors", func() {
						Expect(command.Run(sshOpts)).To(MatchError("the --director flag requires both the --agent-endpoint and --agent-certificate flags to be set"))
					})
				})

				Context("the endpoint and certificate flags are set", func() {

					BeforeEach(func() {
						sshOpts = opts.SSHOpts{
							CreateEnvAuthFlags: opts.CreateEnvAuthFlags{
								TargetDirector: true,
								Endpoint:       "https:///foo:bar@10.0.0.5",
								Certificate:    "some-cert",
							},
							GatewayFlags: opts.GatewayFlags{UUIDGen: uuidGen},
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

						err := command.Run(sshOpts)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("fake-err"))
					})

					itRunsSpecifiedRunnerProperlyWhenCommandGiven := func(runner **fakessh.FakeRunner) {
						Context("when command is provided", func() {
							BeforeEach(func() {
								sshOpts.Command = []string{"cmd", "arg1"}
							})

							It("sets up SSH access, runs SSH command and later cleans up SSH access", func() {
								(*runner).RunStub = func(boshssh.ConnectionOpts, boshdir.SSHResult, []string) error {
									agentClient.EXPECT().CleanUpSSH(gomock.Any()).Times(0)
									return nil
								}
								agentClient.EXPECT().SetUpSSH(gomock.Eq(ExpUsername), mocks.GomegaMock(ContainSubstring("ssh-rsa AAAA"))).
									Times(1)
								agentClient.EXPECT().CleanUpSSH(gomock.Eq(ExpUsername)).
									Times(1)

								Expect(command.Run(sshOpts)).ToNot(HaveOccurred())

								Expect((*runner).RunCallCount()).To(Equal(1))
							})

							It("runs non-interactive SSH", func() {
								agentClient.EXPECT().SetUpSSH(gomock.Any(), gomock.Any()).
									Times(1)
								agentClient.EXPECT().CleanUpSSH(gomock.Any()).
									Times(1)

								Expect(command.Run(sshOpts)).ToNot(HaveOccurred())

								Expect((*runner).RunCallCount()).To(Equal(1))
								Expect(intSSHRunner.RunCallCount()).To(Equal(0))
							})

							It("returns an error if setting up SSH access fails", func() {
								agentClient.EXPECT().SetUpSSH(gomock.Any(), gomock.Any()).
									Return(agentclient.SSHResult{}, errors.New("fake-ssh-err")).
									Times(1)

								err := command.Run(sshOpts)

								Expect(err).To(HaveOccurred())
								Expect(err.Error()).To(ContainSubstring("fake-ssh-err"))
							})

							It("returns an error if generating SSH options fails", func() {
								uuidGen.GenerateError = errors.New("fake-uuid-err")

								err := command.Run(sshOpts)

								Expect(err).To(HaveOccurred())
								Expect(err.Error()).To(ContainSubstring("fake-uuid-err"))
							})

							It("runs non-interactive SSH session with flags, and command", func() {
								result := agentclient.SSHResult{
									Command:       "setup",
									Status:        "success",
									Ip:            "10.0.0.5",
									HostPublicKey: "some-public-key",
								}
								agentClient.EXPECT().SetUpSSH(gomock.Any(), gomock.Any()).
									Return(result, nil).
									Times(1)
								agentClient.EXPECT().CleanUpSSH(gomock.Any()).
									Times(1)

								sshOpts.RawOpts = opts.TrimmedSpaceArgs([]string{"raw1", "raw2"})
								sshOpts.GatewayFlags.Disable = true
								sshOpts.GatewayFlags.Username = "gw-username"
								sshOpts.GatewayFlags.Host = "gw-host"
								sshOpts.GatewayFlags.PrivateKeyPath = "gw-private-key"
								sshOpts.GatewayFlags.SOCKS5Proxy = "socks5"

								Expect(command.Run(sshOpts)).ToNot(HaveOccurred())

								Expect((*runner).RunCallCount()).To(Equal(1))

								runConnOpts, runResult, runCommand := (*runner).RunArgsForCall(0)
								Expect(runConnOpts.RawOpts).To(Equal([]string{"raw1", "raw2", "-o", "StrictHostKeyChecking=yes"}))
								Expect(runConnOpts.PrivateKey).To(ContainSubstring("-----BEGIN RSA PRIVATE KEY-----"))
								Expect(runConnOpts.GatewayDisable).To(Equal(true))
								Expect(runConnOpts.GatewayUsername).To(Equal("gw-username"))
								Expect(runConnOpts.GatewayHost).To(Equal("gw-host"))
								Expect(runConnOpts.GatewayPrivateKeyPath).To(Equal("gw-private-key"))
								Expect(runConnOpts.SOCKS5Proxy).To(Equal("socks5"))
								Expect(runResult).To(Equal(boshdir.SSHResult{Hosts: []boshdir.Host{{Username: ExpUsername, Host: "10.0.0.5", HostPublicKey: "some-public-key", Job: "create-env-vm", IndexOrID: "0"}}}))
								Expect(runCommand).To(Equal([]string{"cmd", "arg1"}))
							})

							It("returns error if non-interactive SSH session errors", func() {
								(*runner).RunReturns(errors.New("fake-err"))
								agentClient.EXPECT().SetUpSSH(gomock.Any(), gomock.Any()).
									Times(1)
								agentClient.EXPECT().CleanUpSSH(gomock.Any()).
									Times(1)

								err := command.Run(sshOpts)

								Expect(err).To(HaveOccurred())
								Expect(err.Error()).To(ContainSubstring("fake-err"))
							})
						})
					}

					Context("when ui is interactive", func() {
						BeforeEach(func() {
							ui.Interactive = true
						})

						itRunsSpecifiedRunnerProperlyWhenCommandGiven(&nonIntSSHRunner)

						It("uses the interactive runner", func() {
							agentClient.EXPECT().SetUpSSH(gomock.Any(), gomock.Any()).
								Times(1)
							agentClient.EXPECT().CleanUpSSH(gomock.Any()).
								Times(1)

							Expect(command.Run(sshOpts)).ToNot(HaveOccurred())

							Expect(intSSHRunner.RunCallCount()).To(Equal(1))
							Expect(nonIntSSHRunner.RunCallCount()).To(Equal(0))
							Expect(resultsSSHRunner.RunCallCount()).To(Equal(0))
						})
					})

					Context("when ui is noninteractive", func() {
						BeforeEach(func() {
							ui.Interactive = false
						})

						itRunsSpecifiedRunnerProperlyWhenCommandGiven(&nonIntSSHRunner)

						It("uses the noninteractive runner", func() {
							agentClient.EXPECT().SetUpSSH(gomock.Any(), gomock.Any()).
								Times(1)
							agentClient.EXPECT().CleanUpSSH(gomock.Any()).
								Times(1)

							Expect(command.Run(sshOpts)).ToNot(HaveOccurred())

							Expect(intSSHRunner.RunCallCount()).To(Equal(0))
							Expect(nonIntSSHRunner.RunCallCount()).To(Equal(1))
							Expect(resultsSSHRunner.RunCallCount()).To(Equal(0))
						})
					})

					Context("when the results option is used", func() {
						BeforeEach(func() {
							sshOpts.Results = true
						})

						itRunsSpecifiedRunnerProperlyWhenCommandGiven(&resultsSSHRunner)

						It("uses the results runner", func() {
							agentClient.EXPECT().SetUpSSH(gomock.Any(), gomock.Any()).
								Times(1)
							agentClient.EXPECT().CleanUpSSH(gomock.Any()).
								Times(1)

							Expect(command.Run(sshOpts)).ToNot(HaveOccurred())

							Expect(intSSHRunner.RunCallCount()).To(Equal(0))
							Expect(nonIntSSHRunner.RunCallCount()).To(Equal(0))
							Expect(resultsSSHRunner.RunCallCount()).To(Equal(1))
						})
					})
				})
			})
		})
	})
})
