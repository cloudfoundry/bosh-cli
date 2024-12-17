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
)

var _ = Describe("SCP", func() {
	const UUID = "8c5ff117-9572-45c5-8564-8bcf076ecafa"
	const ExpUsername = "bosh_8c5ff117957245c"

	Describe("SCPCmd", func() {
		var (
			deployment  *fakedir.FakeDeployment
			uuidGen     *fakeuuid.FakeGenerator
			scpRunner   *fakessh.FakeSCPRunner
			hostBuilder *fakessh.FakeHostBuilder
			command     cmd.SCPCmd
		)

		BeforeEach(func() {
			deployment = &fakedir.FakeDeployment{}
			uuidGen = &fakeuuid.FakeGenerator{}
			scpRunner = &fakessh.FakeSCPRunner{}
			hostBuilder = &fakessh.FakeHostBuilder{}
			command = cmd.NewSCPCmd(scpRunner, hostBuilder)
		})

		Describe("Run", func() {
			var (
				scpOpts opts.SCPOpts
				act     func() error
			)

			BeforeEach(func() {
				scpOpts = opts.SCPOpts{
					GatewayFlags: opts.GatewayFlags{
						UUIDGen: uuidGen,
					},
				}
				uuidGen.GeneratedUUID = UUID

				act = func() error {
					return command.Run(scpOpts, func() (boshdir.Deployment, error) {
						return deployment, nil
					})
				}
			})

			Context("when valid SCP args are provided", func() {
				BeforeEach(func() {
					scpOpts.Args.Paths = []string{"from:file", "/something"}
				})

				It("sets up SSH access, runs SSH command and later cleans up SSH access", func() {
					scpRunner.RunStub = func(boshssh.ConnectionOpts, boshdir.SSHResult, boshssh.SCPArgs) error {
						Expect(deployment.CleanUpSSHCallCount()).To(Equal(0))
						return nil
					}
					Expect(act()).ToNot(HaveOccurred())

					Expect(deployment.SetUpSSHCallCount()).To(Equal(1))
					Expect(scpRunner.RunCallCount()).To(Equal(1))
					Expect(deployment.CleanUpSSHCallCount()).To(Equal(1))

					setupSlug, setupSSHOpts := deployment.SetUpSSHArgsForCall(0)
					Expect(setupSlug).To(Equal(boshdir.NewAllOrInstanceGroupOrInstanceSlug("from", "")))
					Expect(setupSSHOpts.Username).To(Equal(ExpUsername))
					Expect(setupSSHOpts.PublicKey).To(ContainSubstring("ssh-rsa AAAA"))

					slug, sshOpts := deployment.CleanUpSSHArgsForCall(0)
					Expect(slug).To(Equal(setupSlug))
					Expect(sshOpts).To(Equal(setupSSHOpts))
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

				It("runs SCP with flags, and command", func() {
					result := boshdir.SSHResult{Hosts: []boshdir.Host{{Host: "ip1"}}}
					deployment.SetUpSSHReturns(result, nil)

					scpOpts.GatewayFlags.Disable = true
					scpOpts.GatewayFlags.Username = "gw-username"
					scpOpts.GatewayFlags.Host = "gw-host"
					scpOpts.GatewayFlags.PrivateKeyPath = "gw-private-key"
					scpOpts.GatewayFlags.SOCKS5Proxy = "some-proxy"

					Expect(act()).ToNot(HaveOccurred())

					Expect(scpRunner.RunCallCount()).To(Equal(1))

					runConnOpts, runResult, runCommand := scpRunner.RunArgsForCall(0)
					Expect(runConnOpts.PrivateKey).To(ContainSubstring("-----BEGIN RSA PRIVATE KEY-----"))
					Expect(runConnOpts.GatewayDisable).To(Equal(true))
					Expect(runConnOpts.GatewayUsername).To(Equal("gw-username"))
					Expect(runConnOpts.GatewayHost).To(Equal("gw-host"))
					Expect(runConnOpts.GatewayPrivateKeyPath).To(Equal("gw-private-key"))
					Expect(runConnOpts.SOCKS5Proxy).To(Equal("some-proxy"))
					Expect(runResult).To(Equal(boshdir.SSHResult{Hosts: []boshdir.Host{{Host: "ip1"}}}))
					Expect(runCommand).To(Equal(boshssh.NewSCPArgs([]string{"from:file", "/something"}, false)))
				})

				It("sets up SCP to be recursive if recursive flag is set", func() {
					scpOpts.Recursive = true
					Expect(act()).ToNot(HaveOccurred())
					Expect(scpRunner.RunCallCount()).To(Equal(1))

					_, _, runCommand := scpRunner.RunArgsForCall(0)
					Expect(runCommand).To(Equal(boshssh.NewSCPArgs([]string{"from:file", "/something"}, true)))
				})

				It("returns error if SCP errors", func() {
					scpRunner.RunReturns(errors.New("fake-err"))
					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))
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
					scpOpts.Args.Paths = []string{"1.2.3.4:file", "/something"}

					scpOpts.PrivateKey.Bytes = []byte("topsecret")
					scpOpts.Username = "vcap"

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

					expectedSlug, _ := boshdir.NewAllOrInstanceGroupOrInstanceSlugFromString("1.2.3.4")
					Expect(slug).To(Equal(expectedSlug))
					Expect(username).To(Equal(scpOpts.Username))

					Expect(scpRunner.RunCallCount()).To(Equal(1))
					conn, result, _ := scpRunner.RunArgsForCall(0)

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

			Context("when valid SCP args are not provided", func() {
				BeforeEach(func() {
					scpOpts.Args.Paths = []string{"invalid-arg"}
				})

				It("returns an error", func() {
					Expect(act()).To(Equal(errors.New(
						"Missing remote host information in source/destination arguments")))
				})
			})
		})
	})

	Describe("EnvSCPCmd", func() {
		var (
			mockCtrl *gomock.Controller

			agentClientFactory *mockhttpagent.MockAgentClientFactory
			agentClient        *mockagentclient.MockAgentClient
			uuidGen            *fakeuuid.FakeGenerator
			scpRunner          *fakessh.FakeSCPRunner
			command            cmd.EnvSCPCmd
		)

		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())

			agentClient = mockagentclient.NewMockAgentClient(mockCtrl)
			agentClientFactory = mockhttpagent.NewMockAgentClientFactory(mockCtrl)

			uuidGen = &fakeuuid.FakeGenerator{}
			scpRunner = &fakessh.FakeSCPRunner{}
			command = cmd.NewEnvSCPCmd(agentClientFactory, scpRunner)
		})

		AfterEach(func() {
			mockCtrl.Finish()
		})

		Describe("Run", func() {
			var (
				scpOpts opts.SCPOpts
			)

			Context("when private key is provided", func() {
				BeforeEach(func() {
					scpOpts.PrivateKey.Bytes = []byte("topsecret")
				})

				It("errors", func() {
					err := command.Run(scpOpts)

					Expect(err).To(MatchError("the --private-key flag is not supported in combination with the --director flag"))
				})
			})

			Context("when private key is not provided", func() {
				Context("neither the endpoint or certificate flag is set", func() {
					BeforeEach(func() {
						scpOpts = opts.SCPOpts{
							CreateEnvAuthFlags: opts.CreateEnvAuthFlags{
								TargetDirector: true,
							},
						}
					})

					It("errors", func() {
						Expect(command.Run(scpOpts)).To(MatchError("the --director flag requires both the --agent-endpoint and --agent-certificate flags to be set"))
					})
				})

				Context("only the endpoint flag is set", func() {
					BeforeEach(func() {
						scpOpts = opts.SCPOpts{
							CreateEnvAuthFlags: opts.CreateEnvAuthFlags{
								TargetDirector: true,
								Endpoint:       "https:///foo:bar@10.0.0.5",
							},
						}
					})

					It("errors", func() {
						Expect(command.Run(scpOpts)).To(MatchError("the --director flag requires both the --agent-endpoint and --agent-certificate flags to be set"))
					})
				})

				Context("only the certificate flag is set", func() {
					BeforeEach(func() {
						scpOpts = opts.SCPOpts{
							CreateEnvAuthFlags: opts.CreateEnvAuthFlags{
								TargetDirector: true,
								Certificate:    "some-cert",
							},
						}
					})

					It("errors", func() {
						Expect(command.Run(scpOpts)).To(MatchError("the --director flag requires both the --agent-endpoint and --agent-certificate flags to be set"))
					})
				})

				Context("the endpoint and certificate flags are set", func() {
					BeforeEach(func() {
						scpOpts = opts.SCPOpts{
							CreateEnvAuthFlags: opts.CreateEnvAuthFlags{
								TargetDirector: true,
								Endpoint:       "https:///foo:bar@10.0.0.5",
								Certificate:    "some-cert",
							},
							GatewayFlags: opts.GatewayFlags{
								UUIDGen: uuidGen,
							},
						}
						uuidGen.GeneratedUUID = UUID

						agentClientFactory.EXPECT().NewAgentClient(
							gomock.Eq("bosh-cli"),
							gomock.Eq("https:///foo:bar@10.0.0.5"),
							gomock.Eq("some-cert"),
						).Return(agentClient, nil).Times(1)
					})

					Context("when valid SCP args are provided", func() {
						BeforeEach(func() {
							scpOpts.Args.Paths = []string{"from:file", "/something"}
						})

						It("sets up SSH access, runs SSH command and later cleans up SSH access", func() {
							scpRunner.RunStub = func(boshssh.ConnectionOpts, boshdir.SSHResult, boshssh.SCPArgs) error {
								agentClient.EXPECT().CleanUpSSH(gomock.Any()).Times(0)
								return nil
							}
							agentClient.EXPECT().SetUpSSH(gomock.Eq(ExpUsername), mocks.GomegaMock(ContainSubstring("ssh-rsa AAAA"))).
								Times(1)
							agentClient.EXPECT().CleanUpSSH(gomock.Eq(ExpUsername)).
								Times(1)

							Expect(command.Run(scpOpts)).ToNot(HaveOccurred())

							Expect(scpRunner.RunCallCount()).To(Equal(1))
						})

						It("returns an error if setting up SSH access fails", func() {
							agentClient.EXPECT().SetUpSSH(gomock.Any(), gomock.Any()).
								Return(agentclient.SSHResult{}, errors.New("fake-ssh-err")).
								Times(1)

							err := command.Run(scpOpts)

							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("fake-ssh-err"))
						})

						It("returns an error if generating SSH options fails", func() {
							uuidGen.GenerateError = errors.New("fake-uuid-err")

							err := command.Run(scpOpts)

							Expect(err).To(HaveOccurred())
							Expect(err.Error()).To(ContainSubstring("fake-uuid-err"))
						})

						It("runs SCP with flags, and command", func() {
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

							scpOpts.GatewayFlags.Disable = true
							scpOpts.GatewayFlags.Username = "gw-username"
							scpOpts.GatewayFlags.Host = "gw-host"
							scpOpts.GatewayFlags.PrivateKeyPath = "gw-private-key"
							scpOpts.GatewayFlags.SOCKS5Proxy = "some-proxy"

							Expect(command.Run(scpOpts)).ToNot(HaveOccurred())

							Expect(scpRunner.RunCallCount()).To(Equal(1))

							runConnOpts, runResult, runCommand := scpRunner.RunArgsForCall(0)
							Expect(runConnOpts.PrivateKey).To(ContainSubstring("-----BEGIN RSA PRIVATE KEY-----"))
							Expect(runConnOpts.GatewayDisable).To(Equal(true))
							Expect(runConnOpts.GatewayUsername).To(Equal("gw-username"))
							Expect(runConnOpts.GatewayHost).To(Equal("gw-host"))
							Expect(runConnOpts.GatewayPrivateKeyPath).To(Equal("gw-private-key"))
							Expect(runConnOpts.SOCKS5Proxy).To(Equal("some-proxy"))
							Expect(runResult).To(Equal(boshdir.SSHResult{Hosts: []boshdir.Host{{Username: ExpUsername, Host: "10.0.0.5", HostPublicKey: "some-public-key", Job: "create-env-vm", IndexOrID: "0"}}}))
							Expect(runCommand).To(Equal(boshssh.NewSCPArgs([]string{"from:file", "/something"}, false)))
						})

						It("sets up SCP to be recursive if recursive flag is set", func() {
							scpOpts.Recursive = true
							agentClient.EXPECT().SetUpSSH(gomock.Any(), gomock.Any()).
								Times(1)
							agentClient.EXPECT().CleanUpSSH(gomock.Any()).
								Times(1)

							Expect(command.Run(scpOpts)).ToNot(HaveOccurred())
							Expect(scpRunner.RunCallCount()).To(Equal(1))

							_, _, runCommand := scpRunner.RunArgsForCall(0)
							Expect(runCommand).To(Equal(boshssh.NewSCPArgs([]string{"from:file", "/something"}, true)))
						})

						It("returns error if SCP errors", func() {
							agentClient.EXPECT().SetUpSSH(gomock.Any(), gomock.Any()).
								Times(1)
							agentClient.EXPECT().CleanUpSSH(gomock.Any()).
								Times(1)
							scpRunner.RunReturns(errors.New("fake-scp-err"))

							err := command.Run(scpOpts)

							Expect(err).To(MatchError(ContainSubstring("fake-scp-err")))
						})
					})
				})
			})
		})
	})
})
