package config_test

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-utils/uuid/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/config"
)

var _ = Describe("VMRepo", func() {
	var (
		repo                   VMRepo
		deploymentStateService DeploymentStateService
		fs                     *fakesys.FakeFileSystem
		fakeUUIDGenerator      *fakeuuid.FakeGenerator
	)

	BeforeEach(func() {
		logger := boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		fakeUUIDGenerator = &fakeuuid.FakeGenerator{}
		deploymentStateService = NewFileSystemDeploymentStateService(fs, fakeUUIDGenerator, logger, "/fake/path")
		repo = NewVMRepo(deploymentStateService, fakeUUIDGenerator)
	})

	Describe("FindAll", func() {
		Context("when no VMs have been saved", func() {
			It("returns an empty slice", func() {
				records, err := repo.FindAll()
				Expect(err).ToNot(HaveOccurred())
				Expect(records).To(BeEmpty())
			})
		})

		Context("when VMs have been saved", func() {
			BeforeEach(func() {
				_, err := repo.Save("nats", 0, "vm-cid-0", "10.0.0.1", "")
				Expect(err).ToNot(HaveOccurred())
				_, err = repo.Save("nats", 1, "vm-cid-1", "10.0.0.2", "")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns all active VM records", func() {
				records, err := repo.FindAll()
				Expect(err).ToNot(HaveOccurred())
				Expect(records).To(HaveLen(2))
				Expect(records[0].CID).To(Equal("vm-cid-0"))
				Expect(records[0].JobName).To(Equal("nats"))
				Expect(records[0].InstanceID).To(Equal(0))
				Expect(records[0].StaticIP).To(Equal("10.0.0.1"))
				Expect(records[1].CID).To(Equal("vm-cid-1"))
				Expect(records[1].InstanceID).To(Equal(1))
			})
		})
	})

	Describe("Save", func() {
		It("saves a new VM record", func() {
		record, err := repo.Save("nats", 0, "vm-cid-0", "10.0.0.1", "")
		Expect(err).ToNot(HaveOccurred())
		Expect(record.CID).To(Equal("vm-cid-0"))
		Expect(record.JobName).To(Equal("nats"))
		Expect(record.InstanceID).To(Equal(0))
		Expect(record.ID).ToNot(BeEmpty())

			deploymentState, err := deploymentStateService.Load()
			Expect(err).ToNot(HaveOccurred())
			Expect(deploymentState.CurrentVMs).To(HaveLen(1))
		})

		Context("when a pending record (no CID) exists for the same instance", func() {
			BeforeEach(func() {
				// Save then delete to leave a pending record with disk.
				_, err := repo.Save("nats", 0, "vm-cid-0", "10.0.0.1", "")
				Expect(err).ToNot(HaveOccurred())
				err = repo.UpdateCurrentDisk("vm-cid-0", "disk-uuid-1")
				Expect(err).ToNot(HaveOccurred())
				err = repo.Delete("vm-cid-0")
				Expect(err).ToNot(HaveOccurred())
			})

			It("reuses the existing record, preserving CurrentDiskID", func() {
				deploymentState, err := deploymentStateService.Load()
				Expect(err).ToNot(HaveOccurred())
				Expect(deploymentState.CurrentVMs).To(HaveLen(1))
				Expect(deploymentState.CurrentVMs[0].CID).To(BeEmpty())
				Expect(deploymentState.CurrentVMs[0].CurrentDiskID).To(Equal("disk-uuid-1"))

				record, err := repo.Save("nats", 0, "vm-cid-new", "10.0.0.1", "")
				Expect(err).ToNot(HaveOccurred())
				Expect(record.CID).To(Equal("vm-cid-new"))
				Expect(record.CurrentDiskID).To(Equal("disk-uuid-1"))

				// Should not create a new record.
				deploymentState, err = deploymentStateService.Load()
				Expect(err).ToNot(HaveOccurred())
				Expect(deploymentState.CurrentVMs).To(HaveLen(1))
			})
		})
	})

	Describe("UpdateCurrentDisk", func() {
		BeforeEach(func() {
		_, err := repo.Save("nats", 0, "vm-cid-0", "", "")
		Expect(err).ToNot(HaveOccurred())
	})

	It("sets the CurrentDiskID on the VMRecord", func() {
			err := repo.UpdateCurrentDisk("vm-cid-0", "disk-uuid-1")
			Expect(err).ToNot(HaveOccurred())

			deploymentState, err := deploymentStateService.Load()
			Expect(err).ToNot(HaveOccurred())
			Expect(deploymentState.CurrentVMs[0].CurrentDiskID).To(Equal("disk-uuid-1"))
		})

		It("returns an error when VM is not found", func() {
			err := repo.UpdateCurrentDisk("unknown-cid", "disk-uuid-1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("VM record with CID 'unknown-cid' not found"))
		})
	})

	Describe("Delete", func() {
		Context("when the VM has no associated disk", func() {
			BeforeEach(func() {
			_, err := repo.Save("nats", 0, "vm-cid-0", "", "")
			Expect(err).ToNot(HaveOccurred())
		})

		It("removes the record entirely", func() {
				err := repo.Delete("vm-cid-0")
				Expect(err).ToNot(HaveOccurred())

				deploymentState, err := deploymentStateService.Load()
				Expect(err).ToNot(HaveOccurred())
				Expect(deploymentState.CurrentVMs).To(BeEmpty())
			})
		})

		Context("when the VM has an associated disk", func() {
			BeforeEach(func() {
			_, err := repo.Save("nats", 0, "vm-cid-0", "10.0.0.1", "")
			Expect(err).ToNot(HaveOccurred())
			err = repo.UpdateCurrentDisk("vm-cid-0", "disk-uuid-1")
			Expect(err).ToNot(HaveOccurred())
		})

		It("keeps the record (with disk) but clears CID and StaticIP", func() {
				err := repo.Delete("vm-cid-0")
				Expect(err).ToNot(HaveOccurred())

				deploymentState, err := deploymentStateService.Load()
				Expect(err).ToNot(HaveOccurred())
				Expect(deploymentState.CurrentVMs).To(HaveLen(1))
				Expect(deploymentState.CurrentVMs[0].CID).To(BeEmpty())
				Expect(deploymentState.CurrentVMs[0].StaticIP).To(BeEmpty())
				Expect(deploymentState.CurrentVMs[0].CurrentDiskID).To(Equal("disk-uuid-1"))
			})

			It("does not return the (pending) record from FindAll", func() {
				err := repo.Delete("vm-cid-0")
				Expect(err).ToNot(HaveOccurred())

				records, err := repo.FindAll()
				Expect(err).ToNot(HaveOccurred())
				Expect(records).To(BeEmpty())
			})
		})
	})

	Describe("ClearAll", func() {
		BeforeEach(func() {
		_, err := repo.Save("nats", 0, "vm-cid-0", "", "")
		Expect(err).ToNot(HaveOccurred())
		_, err = repo.Save("nats", 1, "vm-cid-1", "", "")
		Expect(err).ToNot(HaveOccurred())
	})

	It("removes all VM records", func() {
			err := repo.ClearAll()
			Expect(err).ToNot(HaveOccurred())

			deploymentState, err := deploymentStateService.Load()
			Expect(err).ToNot(HaveOccurred())
			Expect(deploymentState.CurrentVMs).To(BeNil())
		})
	})
})
