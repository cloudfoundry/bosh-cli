package cmd_test

import (
	"errors"

	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	fakepcap "github.com/cloudfoundry/bosh-cli/v7/pcap/pcapfakes"
)

var _ = Describe("pcap", func() {
	const UUID = "8c5ff117-9572-45c5-8564-8bcf076ecafa"
	const ExpUsername = "bosh_8c5ff117957245c"

	Describe("PcapCmd", func() {
		var (
			deployment *fakedir.FakeDeployment
			uuidGen    *fakeuuid.FakeGenerator
			pcapRunner *fakepcap.FakePcapRunner
			command    cmd.PcapCmd
		)

		BeforeEach(func() {
			deployment = &fakedir.FakeDeployment{}
			uuidGen = &fakeuuid.FakeGenerator{}
			pcapRunner = &fakepcap.FakePcapRunner{}
			command = cmd.NewPcapCmd(deployment, pcapRunner)
		})

		Describe("Run", func() {
			var (
				pcapOpts opts.PcapOpts
				act      func() error
			)

			BeforeEach(func() {
				pcapOpts = opts.PcapOpts{
					GatewayFlags: opts.GatewayFlags{
						UUIDGen: uuidGen,
					},
					SnapLength: 65535,
					Interface:  "eth0",
				}
				uuidGen.GeneratedUUID = UUID

				act = func() error {
					return command.Run(pcapOpts)
				}
			})

			Context("when valid pcap args are provided", func() {
				BeforeEach(func() {
					pcapOpts.Args.Slug = boshdir.AllOrInstanceGroupOrInstanceSlug{}
				})

				It("sets up SSH access, runs SSH command and later cleans up SSH access", func() {
					pcapRunner.RunStub = func(result boshdir.SSHResult, username string, argv string, pcapOpts opts.PcapOpts, privateKey string) error {
						Expect(argv).To(Equal("sudo tcpdump -w - -i eth0 -s 65535"))
						return nil
					}
					Expect(act()).ToNot(HaveOccurred())

					Expect(deployment.SetUpSSHCallCount()).To(Equal(1))
					Expect(deployment.CleanUpSSHCallCount()).To(Equal(1))

					Expect(pcapRunner.RunCallCount()).To(Equal(1))

					_, setupSSHOpts := deployment.SetUpSSHArgsForCall(0)
					Expect(setupSSHOpts.Username).To(Equal(ExpUsername))
					Expect(setupSSHOpts.PublicKey).To(ContainSubstring("ssh-rsa AAAA"))

					_, sshOpts := deployment.CleanUpSSHArgsForCall(0)
					Expect(sshOpts).To(Equal(setupSSHOpts))
				})
				It("returns an error if setting up SSH access fails", func() {
					deployment.SetUpSSHReturns(boshdir.SSHResult{}, errors.New("fake-err"))
					err := act()
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-err"))
				})
				It("provides custom opts, sets up SSH access, runs SSH command and later cleans up SSH access", func() {
					pcapOpts.SnapLength = 300
					pcapOpts.Interface = "any"
					pcapRunner.RunStub = func(result boshdir.SSHResult, username string, argv string, pcapOpts opts.PcapOpts, privateKey string) error {
						Expect(argv).To(Equal("sudo tcpdump -w - -i any -s 300"))
						Expect(deployment.CleanUpSSHCallCount()).To(Equal(0))
						return nil
					}
					Expect(act()).ToNot(HaveOccurred())

					Expect(deployment.SetUpSSHCallCount()).To(Equal(1))
					Expect(deployment.CleanUpSSHCallCount()).To(Equal(1))

					Expect(pcapRunner.RunCallCount()).To(Equal(1))

					_, setupSSHOpts := deployment.SetUpSSHArgsForCall(0)
					Expect(setupSSHOpts.Username).To(Equal(ExpUsername))
					Expect(setupSSHOpts.PublicKey).To(ContainSubstring("ssh-rsa AAAA"))

					_, sshOpts := deployment.CleanUpSSHArgsForCall(0)
					Expect(sshOpts).To(Equal(setupSSHOpts))
				})
			})
		})
	})
})
