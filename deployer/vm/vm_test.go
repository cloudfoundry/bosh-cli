package vm_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"
	fakebmconfig "github.com/cloudfoundry/bosh-micro-cli/config/fakes"
	fakebmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient/fakes"
	fakebmas "github.com/cloudfoundry/bosh-micro-cli/deployer/applyspec/fakes"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/deployer/agentclient"
	bmas "github.com/cloudfoundry/bosh-micro-cli/deployer/applyspec"
	bmdisk "github.com/cloudfoundry/bosh-micro-cli/deployer/disk"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/deployer/stemcell"
	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer/vm"
)

var _ = Describe("VM", func() {
	var (
		vm                         VM
		fakeVMRepo                 *fakebmconfig.FakeVMRepo
		fakeAgentClient            *fakebmagentclient.FakeAgentClient
		fakeCloud                  *fakebmcloud.FakeCloud
		applySpec                  bmstemcell.ApplySpec
		fakeTemplatesSpecGenerator *fakebmas.FakeTemplatesSpecGenerator
		fakeApplySpecFactory       *fakebmas.FakeApplySpecFactory
		deployment                 bmdepl.Deployment
		deploymentJob              bmdepl.Job
		stemcellJob                bmstemcell.Job
		fs                         *fakesys.FakeFileSystem
		logger                     boshlog.Logger
	)

	BeforeEach(func() {
		fakeTemplatesSpecGenerator = fakebmas.NewFakeTemplatesSpecGenerator()
		fakeTemplatesSpecGenerator.SetCreateBehavior(bmas.TemplatesSpec{
			BlobID:            "fake-blob-id",
			ArchiveSha1:       "fake-archive-sha1",
			ConfigurationHash: "fake-configuration-hash",
		}, nil)

		fakeAgentClient = fakebmagentclient.NewFakeAgentClient()
		stemcellJob = bmstemcell.Job{
			Name: "fake-job-name",
			Templates: []bmstemcell.Blob{
				{
					Name:        "first-job-name",
					Version:     "first-job-version",
					SHA1:        "first-job-sha1",
					BlobstoreID: "first-job-blobstore-id",
				},
				{
					Name:        "second-job-name",
					Version:     "second-job-version",
					SHA1:        "second-job-sha1",
					BlobstoreID: "second-job-blobstore-id",
				},
				{
					Name:        "third-job-name",
					Version:     "third-job-version",
					SHA1:        "third-job-sha1",
					BlobstoreID: "third-job-blobstore-id",
				},
			},
		}
		applySpec = bmstemcell.ApplySpec{
			Packages: map[string]bmstemcell.Blob{
				"first-package-name": bmstemcell.Blob{
					Name:        "first-package-name",
					Version:     "first-package-version",
					SHA1:        "first-package-sha1",
					BlobstoreID: "first-package-blobstore-id",
				},
				"second-package-name": bmstemcell.Blob{
					Name:        "second-package-name",
					Version:     "second-package-version",
					SHA1:        "second-package-sha1",
					BlobstoreID: "second-package-blobstore-id",
				},
			},
			Job: stemcellJob,
		}

		deploymentJob = bmdepl.Job{
			Name: "fake-manifest-job-name",
			Templates: []bmdepl.ReleaseJobRef{
				{Name: "first-job-name"},
				{Name: "third-job-name"},
			},
			RawProperties: map[interface{}]interface{}{
				"fake-property-key": "fake-property-value",
			},
			Networks: []bmdepl.JobNetwork{
				{
					Name:      "fake-network-name",
					StaticIPs: []string{"fake-network-ip"},
				},
			},
		}
		deployment = bmdepl.Deployment{
			Name: "fake-deployment-name",
			Jobs: []bmdepl.Job{
				deploymentJob,
			},
			Networks: []bmdepl.Network{
				{
					Name: "fake-network-name",
					Type: "fake-network-type",
				},
			},
		}

		fakeApplySpecFactory = fakebmas.NewFakeApplySpecFactory()

		logger = boshlog.NewLogger(boshlog.LevelNone)
		fs = fakesys.NewFakeFileSystem()
		fakeCloud = fakebmcloud.NewFakeCloud()
		fakeVMRepo = fakebmconfig.NewFakeVMRepo()
		vm = NewVM(
			"fake-vm-cid",
			fakeVMRepo,
			fakeAgentClient,
			fakeCloud,
			fakeTemplatesSpecGenerator,
			fakeApplySpecFactory,
			"fake-mbus-url",
			fs,
			logger,
		)
	})

	Describe("Apply", func() {
		It("stops the agent", func() {
			err := vm.Apply(applySpec, deployment)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeAgentClient.StopCalled).To(BeTrue())
		})

		It("generates templates spec", func() {
			err := vm.Apply(applySpec, deployment)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeTemplatesSpecGenerator.CreateTemplatesSpecInputs).To(ContainElement(fakebmas.CreateTemplatesSpecInput{
				DeploymentJob:  deploymentJob,
				StemcellJob:    stemcellJob,
				DeploymentName: "fake-deployment-name",
				Properties: map[string]interface{}{
					"fake-property-key": "fake-property-value",
				},
				MbusURL: "fake-mbus-url",
			}))
		})

		It("creates apply spec", func() {
			err := vm.Apply(applySpec, deployment)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeApplySpecFactory.CreateInput).To(Equal(
				fakebmas.CreateInput{
					ApplySpec:      applySpec,
					DeploymentName: "fake-deployment-name",
					JobName:        "fake-manifest-job-name",
					NetworksSpec: map[string]interface{}{
						"fake-network-name": map[string]interface{}{
							"type":             "fake-network-type",
							"ip":               "fake-network-ip",
							"cloud_properties": map[string]interface{}{},
						},
					},
					ArchivedTemplatesBlobID: "fake-blob-id",
					ArchivedTemplatesSha1:   "fake-archive-sha1",
					TemplatesDirSha1:        "fake-configuration-hash",
				},
			))
		})

		It("sends apply spec to the agent", func() {
			agentApplySpec := bmas.ApplySpec{
				Deployment: "fake-deployment-name",
			}
			fakeApplySpecFactory.CreateApplySpec = agentApplySpec
			err := vm.Apply(applySpec, deployment)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeAgentClient.ApplyApplySpec).To(Equal(agentApplySpec))
		})

		Context("when creating templates spec fails", func() {
			BeforeEach(func() {
				fakeTemplatesSpecGenerator.CreateErr = errors.New("fake-template-err")
			})

			It("returns an error", func() {
				err := vm.Apply(applySpec, deployment)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-template-err"))
			})
		})

		Context("when sending apply spec to the agent fails", func() {
			BeforeEach(func() {
				fakeAgentClient.ApplyErr = errors.New("fake-agent-apply-err")
			})

			It("returns an error", func() {
				err := vm.Apply(applySpec, deployment)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-agent-apply-err"))
			})
		})

		Context("when stopping an agent fails", func() {
			BeforeEach(func() {
				fakeAgentClient.SetStopBehavior(errors.New("fake-stop-error"))
			})

			It("returns an error", func() {
				err := vm.Apply(applySpec, deployment)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-stop-error"))
			})
		})
	})

	Describe("Start", func() {
		It("starts agent services", func() {
			err := vm.Start()
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeAgentClient.StartCalled).To(BeTrue())
		})

		Context("when starting an agent fails", func() {
			BeforeEach(func() {
				fakeAgentClient.SetStartBehavior(errors.New("fake-start-error"))
			})

			It("returns an error", func() {
				err := vm.Start()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-start-error"))
			})
		})
	})

	Describe("WaitToBeRunning", func() {
		BeforeEach(func() {
			fakeAgentClient.SetGetStateBehavior(bmagentclient.State{JobState: "pending"}, nil)
			fakeAgentClient.SetGetStateBehavior(bmagentclient.State{JobState: "pending"}, nil)
			fakeAgentClient.SetGetStateBehavior(bmagentclient.State{JobState: "running"}, nil)
		})

		It("waits until agent reports state as running", func() {
			err := vm.WaitToBeRunning(5, 0)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeAgentClient.GetStateCalledTimes).To(Equal(3))
		})
	})

	Describe("AttachDisk", func() {
		var disk bmdisk.Disk

		BeforeEach(func() {
			disk = bmdisk.NewDisk("fake-disk-cid")
		})

		It("attaches disk to vm in the cloud", func() {
			err := vm.AttachDisk(disk)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeCloud.AttachDiskInput).To(Equal(fakebmcloud.AttachDiskInput{
				VMCID:   "fake-vm-cid",
				DiskCID: "fake-disk-cid",
			}))
		})

		It("sends mount disk to the agent", func() {
			err := vm.AttachDisk(disk)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeAgentClient.MountDiskCID).To(Equal("fake-disk-cid"))
		})

		Context("when attaching disk to cloud fails", func() {
			BeforeEach(func() {
				fakeCloud.AttachDiskErr = errors.New("fake-attach-error")
			})

			It("returns an error", func() {
				err := vm.AttachDisk(disk)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-attach-error"))
			})
		})

		Context("when mounting disk fails", func() {
			BeforeEach(func() {
				fakeAgentClient.SetMountDiskBehavior(errors.New("fake-mount-error"))
			})

			It("returns an error", func() {
				err := vm.AttachDisk(disk)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-mount-error"))
			})
		})
	})

	Describe("UnmountDisk", func() {
		var disk bmdisk.Disk

		BeforeEach(func() {
			disk = bmdisk.NewDisk("fake-disk-cid")
		})

		It("sends unmount disk to the agent", func() {
			err := vm.UnmountDisk(disk)
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeAgentClient.UnmountDiskCID).To(Equal("fake-disk-cid"))
		})

		Context("when unmounting disk fails", func() {
			BeforeEach(func() {
				fakeAgentClient.SetUnmountDiskBehavior(errors.New("fake-unmount-error"))
			})

			It("returns an error", func() {
				err := vm.UnmountDisk(disk)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-unmount-error"))
			})
		})
	})

	Describe("Stop", func() {
		It("stops agent services", func() {
			err := vm.Stop()
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeAgentClient.StopCalled).To(BeTrue())
		})

		Context("when stopping an agent fails", func() {
			BeforeEach(func() {
				fakeAgentClient.SetStopBehavior(errors.New("fake-stop-error"))
			})

			It("returns an error", func() {
				err := vm.Stop()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-stop-error"))
			})
		})
	})

	Describe("Disks", func() {
		BeforeEach(func() {
			fakeAgentClient.SetListDiskBehavior([]string{"fake-disk-1", "fake-disk-2"}, nil)
		})

		It("returns disks that are reported by the agent", func() {
			disks, err := vm.Disks()
			Expect(err).ToNot(HaveOccurred())
			Expect(disks).To(Equal([]bmdisk.Disk{bmdisk.NewDisk("fake-disk-1"), bmdisk.NewDisk("fake-disk-2")}))
		})

		Context("when listing disks fails", func() {
			BeforeEach(func() {
				fakeAgentClient.SetListDiskBehavior([]string{}, errors.New("fake-list-disk-error"))
			})

			It("returns an error", func() {
				_, err := vm.Disks()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-list-disk-error"))
			})
		})
	})

	Describe("Delete", func() {
		It("deletes vm in the cloud", func() {
			err := vm.Delete()
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeCloud.DeleteVMInput).To(Equal(fakebmcloud.DeleteVMInput{
				VMCID: "fake-vm-cid",
			}))
		})

		It("deletes VM in the vm repo", func() {
			err := vm.Delete()
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeVMRepo.DeleteCurrentCalled).To(BeTrue())
		})

		Context("when deleting vm in the cloud fails", func() {
			BeforeEach(func() {
				fakeCloud.DeleteVMErr = errors.New("fake-delete-vm-error")
			})

			It("returns an error", func() {
				err := vm.Delete()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-delete-vm-error"))
			})
		})
	})
})
