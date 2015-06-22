package devicepathresolver_test

import (
	"time"

	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/infrastructure/devicepathresolver"
	boshsettings "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/settings"
	boshsys "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/system"
	fakesys "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/system/fakes"
)

var _ = Describe("mappedDevicePathResolver", func() {
	var (
		fs           boshsys.FileSystem
		diskSettings boshsettings.DiskSettings
		resolver     DevicePathResolver
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		resolver = NewMappedDevicePathResolver(time.Second, fs)
		diskSettings = boshsettings.DiskSettings{
			Path: "/dev/sda",
		}
	})

	Context("when a matching /dev/xvdX device is found", func() {
		BeforeEach(func() {
			fs.WriteFile("/dev/xvda", []byte{})
			fs.WriteFile("/dev/vda", []byte{})
			fs.WriteFile("/dev/sda", []byte{})
		})

		It("returns the match", func() {
			realPath, timedOut, err := resolver.GetRealDevicePath(diskSettings)
			Expect(err).NotTo(HaveOccurred())
			Expect(timedOut).To(BeFalse())
			Expect(realPath).To(Equal("/dev/xvda"))
		})
	})

	Context("when a matching /dev/vdX device is found", func() {
		BeforeEach(func() {
			fs.WriteFile("/dev/vda", []byte{})
			fs.WriteFile("/dev/sda", []byte{})
		})

		It("returns the match", func() {
			realPath, timedOut, err := resolver.GetRealDevicePath(diskSettings)
			Expect(err).NotTo(HaveOccurred())
			Expect(timedOut).To(BeFalse())
			Expect(realPath).To(Equal("/dev/vda"))
		})
	})

	Context("when a matching /dev/sdX device is found", func() {
		BeforeEach(func() {
			fs.WriteFile("/dev/sda", []byte{})
		})

		It("returns the match", func() {
			realPath, timedOut, err := resolver.GetRealDevicePath(diskSettings)
			Expect(err).NotTo(HaveOccurred())
			Expect(timedOut).To(BeFalse())
			Expect(realPath).To(Equal("/dev/sda"))
		})
	})

	Context("when no matching device is found the first time", func() {
		Context("when the timeout has not expired", func() {
			BeforeEach(func() {
				time.AfterFunc(time.Second, func() {
					fs.WriteFile("/dev/xvda", []byte{})
				})
			})

			It("returns the match", func() {
				realPath, timedOut, err := resolver.GetRealDevicePath(diskSettings)
				Expect(err).NotTo(HaveOccurred())
				Expect(timedOut).To(BeFalse())
				Expect(realPath).To(Equal("/dev/xvda"))
			})
		})

		Context("when the timeout has expired", func() {
			It("errs", func() {
				_, timedOut, err := resolver.GetRealDevicePath(diskSettings)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Timed out getting real device path for /dev/sda"))
				Expect(timedOut).To(BeTrue())
			})
		})
	})

	Context("when an invalid device name is passed in", func() {
		It("panics", func() {
			Expect(func() {
				diskSettings = boshsettings.DiskSettings{
					Path: "not even a device",
				}
				resolver.GetRealDevicePath(diskSettings)
			}).To(Panic())
		})
	})
})
