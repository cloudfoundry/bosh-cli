package disk_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmconfig "github.com/cloudfoundry/bosh-micro-cli/config"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
)

var _ = Describe("Manager", func() {
	var (
		manager           Manager
		fakeCloud         *fakebmcloud.FakeCloud
		fakeFs            *fakesys.FakeFileSystem
		fakeUUIDGenerator *fakeuuid.FakeGenerator
		diskRepo          bmconfig.DiskRepo
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fakeFs = fakesys.NewFakeFileSystem()
		configService := bmconfig.NewFileSystemDeploymentConfigService("/fake/path", fakeFs, logger)
		fakeUUIDGenerator = &fakeuuid.FakeGenerator{}
		diskRepo = bmconfig.NewDiskRepo(configService, fakeUUIDGenerator)
		managerFactory := NewManagerFactory(diskRepo, logger)
		fakeCloud = fakebmcloud.NewFakeCloud()
		manager = managerFactory.NewManager(fakeCloud)
		fakeUUIDGenerator.GeneratedUuid = "fake-uuid"
	})

	Describe("Create", func() {
		var (
			diskPool bmdepl.DiskPool
		)

		BeforeEach(func() {

			diskPool = bmdepl.DiskPool{
				Name: "fake-disk-pool-name",
				Size: 1024,
				RawCloudProperties: map[interface{}]interface{}{
					"fake-cloud-property-key": "fake-cloud-property-value",
				},
			}
		})

		Context("when creating disk succeeds", func() {
			BeforeEach(func() {
				fakeCloud.CreateDiskCID = "fake-disk-cid"
			})

			It("returns a disk", func() {
				disk, err := manager.Create(diskPool, "fake-vm-cid")
				Expect(err).ToNot(HaveOccurred())
				Expect(disk.CID()).To(Equal("fake-disk-cid"))
			})

			It("saves the disk record", func() {
				_, err := manager.Create(diskPool, "fake-vm-cid")
				Expect(err).ToNot(HaveOccurred())

				diskRecord, found, err := diskRepo.FindCurrent()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())

				Expect(diskRecord).To(Equal(bmconfig.DiskRecord{
					ID:  "fake-uuid",
					CID: "fake-disk-cid",
				}))
			})
		})

		Context("when creating disk fails", func() {
			BeforeEach(func() {
				fakeCloud.CreateDiskErr = errors.New("fake-create-error")
			})

			It("returns an error", func() {
				_, err := manager.Create(diskPool, "fake-vm-cid")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-create-error"))
			})
		})

		Context("when updating disk record fails", func() {
			BeforeEach(func() {
				fakeFs.WriteToFileError = errors.New("fake-write-error")
			})

			It("returns an error", func() {
				_, err := manager.Create(diskPool, "fake-vm-cid")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-write-error"))
			})
		})
	})

	Describe("Find", func() {
		Context("when disk already exists in disk repo", func() {
			BeforeEach(func() {
				diskRecord, err := diskRepo.Save("fake-existing-disk-cid")
				Expect(err).ToNot(HaveOccurred())

				err = diskRepo.UpdateCurrent(diskRecord.ID)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the existing disk", func() {
				disk, found, err := manager.Find()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(disk.CID()).To(Equal("fake-existing-disk-cid"))
			})
		})

		Context("when disk does not exists in disk repo", func() {
			It("returns false", func() {
				_, found, err := manager.Find()
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(BeFalse())
			})
		})

		Context("when reading disk repo fails", func() {
			BeforeEach(func() {
				fakeFs.WriteFileString("/fake/path", "{}")
				fakeFs.ReadFileError = errors.New("fake-read-error")
			})

			It("returns an error", func() {
				_, found, err := manager.Find()
				Expect(found).To(BeFalse())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-read-error"))
			})
		})
	})
})
