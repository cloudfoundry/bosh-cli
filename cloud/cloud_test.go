package cloud_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakebmcloud "github.com/cloudfoundry/bosh-micro-cli/cloud/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/cloud"
)

var _ = Describe("Cloud", func() {
	var (
		cloud            Cloud
		fakeCPICmdRunner *fakebmcloud.FakeCPICmdRunner
		deploymentUUID   string
	)

	BeforeEach(func() {
		fakeCPICmdRunner = fakebmcloud.NewFakeCPICmdRunner()
		deploymentUUID = "fake-uuid"
		logger := boshlog.NewLogger(boshlog.LevelNone)
		cloud = NewCloud(fakeCPICmdRunner, deploymentUUID, logger)
	})

	Describe("CreateStemcell", func() {
		var (
			stemcellImagePath string
			cloudProperties   map[string]interface{}
		)

		BeforeEach(func() {
			stemcellImagePath = "/stemcell/path"
			cloudProperties = map[string]interface{}{
				"fake-key": "fake-value",
			}
		})

		Context("when the cpi successfully creates the stemcell", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunCmdOutput = CmdOutput{
					Result: "fake-cid",
				}
			})

			It("executes the cpi job script with stemcell image path & cloud_properties", func() {
				_, err := cloud.CreateStemcell(cloudProperties, stemcellImagePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.RunInputs).To(HaveLen(1))
				Expect(fakeCPICmdRunner.RunInputs[0]).To(Equal(fakebmcloud.RunInput{
					Method: "create_stemcell",
					Arguments: []interface{}{
						stemcellImagePath,
						cloudProperties,
					},
				}))
			})

			It("returns the cid returned from executing the cpi script", func() {
				cid, err := cloud.CreateStemcell(cloudProperties, stemcellImagePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(cid).To(Equal("fake-cid"))
			})
		})

		Context("when the result is of an unexpected type", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunCmdOutput = CmdOutput{
					Result: 1,
				}
			})

			It("returns an error", func() {
				_, err := cloud.CreateStemcell(cloudProperties, stemcellImagePath)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unexpected external CPI command result: '1'"))
			})
		})

		Context("when the cpi returns an error", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunErr = errors.New("fake-run-error")
			})

			It("returns an error", func() {
				_, err := cloud.CreateStemcell(cloudProperties, stemcellImagePath)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-error"))
			})
		})
	})

	Describe("CreateVM", func() {
		var (
			stemcellCID     string
			cloudProperties map[string]interface{}
			networksSpec    map[string]interface{}
			env             map[string]interface{}
		)

		BeforeEach(func() {
			stemcellCID = "fake-stemcell-cid"
			networksSpec = map[string]interface{}{
				"bosh": map[string]interface{}{
					"type": "dynamic",
					"cloud_properties": map[string]interface{}{
						"a": "b",
					},
				},
			}
			cloudProperties = map[string]interface{}{
				"fake-cloud-property-key": "fake-cloud-property-value",
			}
			env = map[string]interface{}{
				"fake-env-key": "fake-env-value",
			}
		})

		Context("when the cpi successfully creates the vm", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunCmdOutput = CmdOutput{
					Result: "fake-vm-cid",
				}
			})

			It("executes the cpi job script with the director UUID and stemcell CID", func() {
				_, err := cloud.CreateVM(stemcellCID, cloudProperties, networksSpec, env)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.RunInputs).To(HaveLen(1))
				Expect(fakeCPICmdRunner.RunInputs[0]).To(Equal(fakebmcloud.RunInput{
					Method: "create_vm",
					Arguments: []interface{}{
						deploymentUUID,
						stemcellCID,
						cloudProperties,
						networksSpec,
						[]interface{}{},
						env,
					},
				}))
			})

			It("returns the cid returned from executing the cpi script", func() {
				cid, err := cloud.CreateVM(stemcellCID, cloudProperties, networksSpec, env)
				Expect(err).NotTo(HaveOccurred())
				Expect(cid).To(Equal("fake-vm-cid"))
			})
		})

		Context("when the result is of an unexpected type", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunCmdOutput = CmdOutput{
					Result: 1,
				}
			})

			It("returns an error", func() {
				_, err := cloud.CreateVM(stemcellCID, cloudProperties, networksSpec, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unexpected external CPI command result: '1'"))
			})
		})

		Context("when the cpi returns an error", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunErr = errors.New("fake-run-error")
			})

			It("returns an error", func() {
				_, err := cloud.CreateVM(stemcellCID, cloudProperties, networksSpec, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-error"))
			})
		})
	})

	Describe("CreateDisk", func() {
		var (
			size            int
			cloudProperties map[string]interface{}
			instanceID      string
		)

		BeforeEach(func() {
			size = 1024
			cloudProperties = map[string]interface{}{
				"fake-cloud-property-key": "fake-cloud-property-value",
			}
			instanceID = "fake-instance-id"
		})

		Context("when the cpi successfully creates the disk", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunCmdOutput = CmdOutput{
					Result: "fake-disk-cid",
				}
			})

			It("executes the cpi job script with the correct arguments", func() {
				_, err := cloud.CreateDisk(size, cloudProperties, instanceID)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.RunInputs).To(HaveLen(1))
				Expect(fakeCPICmdRunner.RunInputs[0]).To(Equal(fakebmcloud.RunInput{
					Method: "create_disk",
					Arguments: []interface{}{
						size,
						cloudProperties,
						instanceID,
					},
				}))
			})

			It("returns the cid returned from executing the cpi script", func() {
				cid, err := cloud.CreateDisk(size, cloudProperties, instanceID)
				Expect(err).NotTo(HaveOccurred())
				Expect(cid).To(Equal("fake-disk-cid"))
			})
		})

		Context("when the result is of an unexpected type", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunCmdOutput = CmdOutput{
					Result: 1,
				}
			})

			It("returns an error", func() {
				_, err := cloud.CreateDisk(size, cloudProperties, instanceID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unexpected external CPI command result: '1'"))
			})
		})

		Context("when the cpi returns an error", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunErr = errors.New("fake-run-error")
			})

			It("returns an error", func() {
				_, err := cloud.CreateDisk(size, cloudProperties, instanceID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-error"))
			})
		})
	})

	Describe("AttachDisk", func() {
		Context("when the cpi successfully creates the disk", func() {
			It("executes the cpi job script with the correct arguments", func() {
				err := cloud.AttachDisk("fake-vm-cid", "fake-disk-cid")
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.RunInputs).To(HaveLen(1))
				Expect(fakeCPICmdRunner.RunInputs[0]).To(Equal(fakebmcloud.RunInput{
					Method: "attach_disk",
					Arguments: []interface{}{
						"fake-vm-cid",
						"fake-disk-cid",
					},
				}))
			})
		})

		Context("when the cpi returns an error", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunErr = errors.New("fake-attach-error")
			})

			It("returns an error", func() {
				err := cloud.AttachDisk("fake-vm-cid", "fake-disk-cid")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-attach-error"))
			})
		})
	})

	Describe("DeleteVM", func() {
		Context("when the cpi successfully deletes vm", func() {
			It("executes the cpi job script with the correct arguments", func() {
				err := cloud.DeleteVM("fake-vm-cid")
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.RunInputs).To(HaveLen(1))
				Expect(fakeCPICmdRunner.RunInputs[0]).To(Equal(fakebmcloud.RunInput{
					Method: "delete_vm",
					Arguments: []interface{}{
						"fake-vm-cid",
					},
				}))
			})
		})

		Context("when the cpi returns an error", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunErr = errors.New("fake-delete-error")
			})

			It("returns an error", func() {
				err := cloud.DeleteVM("fake-vm-cid")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-delete-error"))
			})
		})
	})
})
