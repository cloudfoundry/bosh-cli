package disk_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	. "github.com/cloudfoundry/bosh-agent/platform/disk"
	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
)

var _ = Describe("rootDevicePartitioner", func() {
	var (
		fakeCmdRunner *fakesys.FakeCmdRunner
		partitioner   Partitioner
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fakeCmdRunner = fakesys.NewFakeCmdRunner()
		partitioner = NewRootDevicePartitioner(logger, fakeCmdRunner, 1)
	})

	Describe("Partition", func() {
		Context("when the desired partitions do not exist", func() {
			BeforeEach(func() {
				fakeCmdRunner.AddCmdResult(
					"parted -m /dev/sda unit B print",
					fakesys.FakeCmdResult{
						Stdout: `BYT;
/dev/sda:129B:virtblk:512:512:msdos:Virtio Block Device;
1:1B:32B:32B:ext4::;
`,
					},
				)
			})

			It("creates partitions using parted", func() {
				partitions := []Partition{
					{SizeInBytes: 32},
					{SizeInBytes: 64},
				}

				err := partitioner.Partition("/dev/sda", partitions)
				Expect(err).ToNot(HaveOccurred())

				Expect(len(fakeCmdRunner.RunCommands)).To(Equal(3))
				Expect(fakeCmdRunner.RunCommands).To(ContainElement([]string{"parted", "-m", "/dev/sda", "unit", "B", "print"}))
				Expect(fakeCmdRunner.RunCommands).To(ContainElement([]string{"parted", "-s", "/dev/sda", "unit", "B", "mkpart", "primary", "33", "64"}))
				Expect(fakeCmdRunner.RunCommands).To(ContainElement([]string{"parted", "-s", "/dev/sda", "unit", "B", "mkpart", "primary", "65", "128"}))
			})

			Context("when partitioning fails", func() {
				BeforeEach(func() {
					fakeCmdRunner.AddCmdResult(
						"parted -s /dev/sda unit B mkpart primary 33 64",
						fakesys.FakeCmdResult{Error: errors.New("fake-parted-error")},
					)
				})

				It("returns error", func() {
					partitions := []Partition{
						{SizeInBytes: 32},
					}

					err := partitioner.Partition("/dev/sda", partitions)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Partitioning disk `/dev/sda'"))
					Expect(err.Error()).To(ContainSubstring("fake-parted-error"))
				})
			})
		})

		Context("when getting existing partitions fails", func() {
			BeforeEach(func() {
				fakeCmdRunner.AddCmdResult(
					"parted -m /dev/sda unit B print",
					fakesys.FakeCmdResult{Error: errors.New("fake-parted-error")},
				)
			})

			It("returns error", func() {
				partitions := []Partition{
					{SizeInBytes: 32},
				}

				err := partitioner.Partition("/dev/sda", partitions)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Getting existing partitions of `/dev/sda'"))
				Expect(err.Error()).To(ContainSubstring("Running parted print on `/dev/sda'"))
				Expect(err.Error()).To(ContainSubstring("fake-parted-error"))
			})
		})

		Context("when partitions already match", func() {
			BeforeEach(func() {
				fakeCmdRunner.AddCmdResult(
					"parted -m /dev/sda unit B print",
					fakesys.FakeCmdResult{
						Stdout: `BYT;
/dev/sda:128B:virtblk:512:512:msdos:Virtio Block Device;
1:1B:32B:32B:ext4::;
2:33B:64B:32B:ext4::;
`,
					},
				)
			})

			It("does not partition", func() {
				partitions := []Partition{{SizeInBytes: 32}}

				err := partitioner.Partition("/dev/sda", partitions)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(fakeCmdRunner.RunCommands)).To(Equal(1))
				Expect(fakeCmdRunner.RunCommands).To(ContainElement([]string{"parted", "-m", "/dev/sda", "unit", "B", "print"}))
			})
		})

		Context("when partitions are within delta", func() {
			BeforeEach(func() {
				fakeCmdRunner.AddCmdResult(
					"parted -m /dev/sda unit B print",
					fakesys.FakeCmdResult{
						Stdout: `BYT;
/dev/sda:128B:virtblk:512:512:msdos:Virtio Block Device;
1:1B:31B:31B:ext4::;
2:32B:64B:33B:ext4::;
3:65B:125B:61B:ext4::;
`,
					},
				)
			})

			It("does not partition", func() {
				partitions := []Partition{
					{SizeInBytes: 32},
					{SizeInBytes: 62},
				}

				err := partitioner.Partition("/dev/sda", partitions)
				Expect(err).ToNot(HaveOccurred())
				Expect(len(fakeCmdRunner.RunCommands)).To(Equal(1))
				Expect(fakeCmdRunner.RunCommands).To(ContainElement([]string{"parted", "-m", "/dev/sda", "unit", "B", "print"}))
			})
		})

		Context("when partition in the middle does not match", func() {
			BeforeEach(func() {
				fakeCmdRunner.AddCmdResult(
					"parted -m /dev/sda unit B print",
					fakesys.FakeCmdResult{
						Stdout: `BYT;
/dev/sda:128B:virtblk:512:512:msdos:Virtio Block Device;
1:1B:32B:32B:ext4::;
2:33B:47B:15B:ext4::;
3:48B:79B:32B:ext4::;
4:80B:111B:32B:ext4::;
5:112B:119B:8B:ext4::;
`,
					},
				)
			})

			It("returns an error", func() {
				partitions := []Partition{
					{SizeInBytes: 16},
					{SizeInBytes: 16},
					{SizeInBytes: 32},
				}

				err := partitioner.Partition("/dev/sda", partitions)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Found 4 unexpected partitions on `/dev/sda'"))
				Expect(fakeCmdRunner.RunCommands).To(Equal([][]string{
					{"parted", "-m", "/dev/sda", "unit", "B", "print"},
				}))
			})
		})

		Context("when the first partition is missing", func() {
			BeforeEach(func() {
				fakeCmdRunner.AddCmdResult(
					"parted -m /dev/sda unit B print",
					fakesys.FakeCmdResult{
						Stdout: `BYT;
/dev/sda:128B:virtblk:512:512:msdos:Virtio Block Device;
`,
					},
				)
			})

			It("returns an error", func() {
				partitions := []Partition{
					{SizeInBytes: 32},
				}

				err := partitioner.Partition("/dev/sda", partitions)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Missing first partition on `/dev/sda'"))
				Expect(len(fakeCmdRunner.RunCommands)).To(Equal(1))
				Expect(fakeCmdRunner.RunCommands).To(ContainElement([]string{"parted", "-m", "/dev/sda", "unit", "B", "print"}))
			})
		})

		Context("when checking existing partitions does not return any result", func() {
			BeforeEach(func() {
				fakeCmdRunner.AddCmdResult(
					"parted -m /dev/sda unit B print",
					fakesys.FakeCmdResult{
						Stdout: "",
					},
				)
			})

			It("returns an error", func() {
				partitions := []Partition{
					{SizeInBytes: 32},
				}

				err := partitioner.Partition("/dev/sda", partitions)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Parsing existing partitions of `/dev/sda'"))
				Expect(len(fakeCmdRunner.RunCommands)).To(Equal(1))
				Expect(fakeCmdRunner.RunCommands).To(ContainElement([]string{"parted", "-m", "/dev/sda", "unit", "B", "print"}))
			})
		})

		Context("when checking existing partitions does not return any result", func() {
			BeforeEach(func() {
				fakeCmdRunner.AddCmdResult(
					"parted -m /dev/sda unit B print",
					fakesys.FakeCmdResult{
						Stdout: `BYT;
/dev/sda:128B:virtblk:512:512:msdos:Virtio Block Device;
1:1B:32B:32B:ext4::;
2:0.2B:65B:32B:ext4::;
`,
					},
				)
			})

			It("returns an error", func() {
				partitions := []Partition{
					{SizeInBytes: 32},
				}

				err := partitioner.Partition("/dev/sda", partitions)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Parsing existing partitions of `/dev/sda'"))
				Expect(len(fakeCmdRunner.RunCommands)).To(Equal(1))
				Expect(fakeCmdRunner.RunCommands).To(ContainElement([]string{"parted", "-m", "/dev/sda", "unit", "B", "print"}))
			})
		})
	})

	Describe("GetDeviceSizeInBytes", func() {
		Context("when getting disk partition information succeeds", func() {
			BeforeEach(func() {
				fakeCmdRunner.AddCmdResult(
					"parted -m /dev/sda unit B print",
					fakesys.FakeCmdResult{
						Stdout: `BYT;
/dev/sda:128B:virtblk:512:512:msdos:Virtio Block Device;
1:15B:31B:17B:ext4::;
2:32B:54B:23B:ext4::;
`,
					},
				)
			})

			It("returns the size of the device", func() {
				size, err := partitioner.GetDeviceSizeInBytes("/dev/sda")
				Expect(err).ToNot(HaveOccurred())
				Expect(size).To(Equal(uint64(96)))
			})
		})

		Context("when getting disk partition information fails", func() {
			BeforeEach(func() {
				fakeCmdRunner.AddCmdResult(
					"parted -m /dev/sda unit B print",
					fakesys.FakeCmdResult{
						Error: errors.New("fake-parted-error"),
					},
				)
			})

			It("returns an error", func() {
				size, err := partitioner.GetDeviceSizeInBytes("/dev/sda")
				Expect(err).To(HaveOccurred())
				Expect(size).To(Equal(uint64(0)))
				Expect(err.Error()).To(ContainSubstring("fake-parted-error"))
			})
		})

		Context("when parsing parted result fails", func() {
			BeforeEach(func() {
				fakeCmdRunner.AddCmdResult(
					"parted -m /dev/sda unit B print",
					fakesys.FakeCmdResult{
						Stdout: ``,
					},
				)
			})

			It("returns an error", func() {
				size, err := partitioner.GetDeviceSizeInBytes("/dev/sda")
				Expect(err).To(HaveOccurred())
				Expect(size).To(Equal(uint64(0)))
				Expect(err.Error()).To(ContainSubstring("Getting remaining size of `/dev/sda'"))
			})
		})
	})
})
