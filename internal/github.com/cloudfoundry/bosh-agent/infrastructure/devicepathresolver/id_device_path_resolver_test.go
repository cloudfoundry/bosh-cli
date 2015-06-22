package devicepathresolver_test

import (
	"errors"
	"os"
	"time"

	fakeudev "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/platform/udevdevice/fakes"
	boshsettings "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/settings"
	fakesys "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/system/fakes"

	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry/bosh-init/internal/github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-agent/infrastructure/devicepathresolver"
)

var _ = Describe("IDDevicePathResolver", func() {
	var (
		fs           *fakesys.FakeFileSystem
		udev         *fakeudev.FakeUdevDevice
		diskSettings boshsettings.DiskSettings
		pathResolver DevicePathResolver
	)

	BeforeEach(func() {
		udev = fakeudev.NewFakeUdevDevice()
		fs = fakesys.NewFakeFileSystem()
		pathResolver = NewIDDevicePathResolver(500*time.Millisecond, udev, fs)
		diskSettings = boshsettings.DiskSettings{
			ID: "fake-disk-id-include-truncate",
		}
	})

	Describe("GetRealDevicePath", func() {
		It("refreshes udev", func() {
			pathResolver.GetRealDevicePath(diskSettings)
			Expect(udev.Triggered).To(Equal(true))
			Expect(udev.Settled).To(Equal(true))
		})

		Context("when path exists", func() {
			BeforeEach(func() {
				err := fs.MkdirAll("fake-device-path", os.FileMode(0750))
				Expect(err).ToNot(HaveOccurred())

				err = fs.Symlink("fake-device-path", "/dev/disk/by-id/virtio-fake-disk-id-include")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the path ", func() {
				path, timeout, err := pathResolver.GetRealDevicePath(diskSettings)
				Expect(err).ToNot(HaveOccurred())

				Expect(path).To(Equal("fake-device-path"))
				Expect(timeout).To(BeFalse())
			})
		})

		Context("when path does not exist", func() {
			BeforeEach(func() {
				err := fs.Symlink("fake-device-path", "/dev/disk/by-id/virtio-fake-disk-id-include")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				_, _, err := pathResolver.GetRealDevicePath(diskSettings)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when symlink does not exist", func() {
			It("returns an error", func() {
				_, _, err := pathResolver.GetRealDevicePath(diskSettings)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when no matching device is found the first time", func() {
			Context("when the timeout has not expired", func() {
				BeforeEach(func() {
					time.AfterFunc(100*time.Millisecond, func() {
						err := fs.MkdirAll("fake-device-path", os.FileMode(0750))
						Expect(err).ToNot(HaveOccurred())

						err = fs.Symlink("fake-device-path", "/dev/disk/by-id/virtio-fake-disk-id-include")
						Expect(err).ToNot(HaveOccurred())
					})
				})

				It("returns the real path", func() {
					path, timeout, err := pathResolver.GetRealDevicePath(diskSettings)
					Expect(err).ToNot(HaveOccurred())

					Expect(path).To(Equal("fake-device-path"))
					Expect(timeout).To(BeFalse())
				})
			})
		})

		Context("when triggering udev fails", func() {
			BeforeEach(func() {
				udev.TriggerErr = errors.New("fake-udev-trigger-error")
			})

			It("returns an error", func() {
				_, timeout, err := pathResolver.GetRealDevicePath(diskSettings)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-udev-trigger-error"))
				Expect(timeout).To(BeFalse())
			})
		})

		Context("when settling udev fails", func() {
			BeforeEach(func() {
				udev.SettleErr = errors.New("fake-udev-settle-error")
			})

			It("returns an error", func() {
				_, timeout, err := pathResolver.GetRealDevicePath(diskSettings)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-udev-settle-error"))
				Expect(timeout).To(BeFalse())
			})
		})

		Context("when id is empty", func() {
			BeforeEach(func() {
				diskSettings = boshsettings.DiskSettings{}
			})

			It("returns an error", func() {
				_, timeout, err := pathResolver.GetRealDevicePath(diskSettings)
				Expect(err).To(HaveOccurred())
				Expect(timeout).To(BeFalse())
			})
		})

		Context("when id is not the correct format", func() {
			BeforeEach(func() {
				diskSettings = boshsettings.DiskSettings{
					ID: "too-short",
				}
			})

			It("returns an error", func() {
				_, timeout, err := pathResolver.GetRealDevicePath(diskSettings)
				Expect(err).To(HaveOccurred())
				Expect(timeout).To(BeFalse())
			})
		})
	})
})
