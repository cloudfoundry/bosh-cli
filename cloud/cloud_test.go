package cloud_test

import (
	"errors"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cloud/cloudfakes"

	. "github.com/cloudfoundry/bosh-cli/v7/cloud"
)

var _ = Describe("Cloud", func() {
	var (
		cloud               Cloud
		expectedContext     CmdContext
		fakeCPICmdRunner    *cloudfakes.FakeCPICmdRunner
		logger              boshlog.Logger
		stemcellApiVersion  = 1
		cpiApiVersion       = 1
		infoResult          map[string]interface{}
		infoResultWithApiV2 map[string]interface{}
	)

	BeforeEach(func() {
		fakeCPICmdRunner = &cloudfakes.FakeCPICmdRunner{}
		logger = boshlog.NewLogger(boshlog.LevelNone)
		infoResult = map[string]interface{}{
			"stemcell_formats": []interface{}{"aws-raw", "aws-light"},
		}
		infoResultWithApiV2 = map[string]interface{}{
			"stemcell_formats": []interface{}{"aws-raw", "aws-light"},
			"api_version":      float64(2),
		}
	})

	JustBeforeEach(func() {
		expectedContext = CmdContext{DirectorID: "fake-director-id", Vm: &VM{Stemcell: &Stemcell{ApiVersion: stemcellApiVersion}}}
		cloud = NewCloud(fakeCPICmdRunner, "fake-director-id", stemcellApiVersion, logger)
	})

	var itHandlesCPIErrors = func(method string, exec func() error) {
		It("returns a cloud.Error when the CPI command returns an error", func() {
			fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{Result: infoResult}, nil)
			fakeCPICmdRunner.RunReturnsOnCall(1, CmdOutput{Error: &CmdError{
				Type:    "Bosh::Cloud::CloudError",
				Message: "fake-cpi-error-msg",
			}}, nil)

			err := exec()
			Expect(err).To(HaveOccurred())

			cpiError, ok := err.(Error)
			Expect(ok).To(BeTrue(), "Expected %s to implement the Error interface", cpiError)
			Expect(cpiError.Method()).To(Equal(method))
			Expect(cpiError.Type()).To(Equal("Bosh::Cloud::CloudError"))
			Expect(cpiError.Message()).To(Equal("fake-cpi-error-msg"))
			Expect(err.Error()).To(ContainSubstring("Bosh::Cloud::CloudError"))
			Expect(err.Error()).To(ContainSubstring("fake-cpi-error-msg"))
		})
	}

	Describe("Info", func() {
		Context("when the stemcell version is 2", func() {
			BeforeEach(func() {
				stemcellApiVersion = 2
			})

			It("return info based on cpi", func() {
				infoParsed := CpiInfo{
					StemcellFormats: []string{"aws-raw", "aws-light"},
					ApiVersion:      2,
				}
				fakeCPICmdRunner.RunReturns(CmdOutput{Result: infoResultWithApiV2}, nil)
				cpiInfo, err := cloud.Info()
				Expect(cpiInfo).To(Equal(infoParsed))
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeCPICmdRunner.RunCallCount()).To(Equal(1))
				context, method, apiVersion, args := fakeCPICmdRunner.RunArgsForCall(0)
				Expect(context).To(Equal(expectedContext))
				Expect(method).To(Equal("info"))
				Expect(apiVersion).To(Equal(1))
				Expect(args).To(BeNil())
			})

			It("uses a default cpi api version if an old cpi does not have api version", func() {
				infoParsed := CpiInfo{
					StemcellFormats: []string{"aws-raw", "aws-light"},
					ApiVersion:      1,
				}
				fakeCPICmdRunner.RunReturns(CmdOutput{Result: infoResult}, nil)
				cpiInfo, err := cloud.Info()
				Expect(cpiInfo).To(Equal(infoParsed))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("when the cpi command execution fails", func() {
				BeforeEach(func() {
					fakeCPICmdRunner.RunReturns(CmdOutput{}, errors.New("info"))
				})

				It("returns an error", func() {
					_, err := cloud.Info()
					Expect(err).To(HaveOccurred())
				})
			})

			Context("when the cpi version is > 2", func() {
				It("should return MAX supported version by CLI", func() {
					infoResult = map[string]interface{}{
						"stemcell_formats": []interface{}{"aws-raw", "aws-light"},
						"api_version":      float64(42),
					}
					infoParsed := CpiInfo{
						StemcellFormats: []string{"aws-raw", "aws-light"},
						ApiVersion:      2,
					}
					fakeCPICmdRunner.RunReturns(CmdOutput{Result: infoResult}, nil)
					cpiInfo, err := cloud.Info()
					Expect(err).ToNot(HaveOccurred())
					Expect(cpiInfo).To(Equal(infoParsed))
				})
			})

			Context("when info return unexpected format result", func() {
				Context("when api_version is not a number format", func() {
					BeforeEach(func() {
						infoResultWithApiV2 = map[string]interface{}{
							"stemcell_formats": []interface{}{"aws-raw", "aws-light"},
							"api_version":      "57",
						}

						fakeCPICmdRunner.RunReturns(CmdOutput{Result: infoResultWithApiV2}, nil)
					})

					It("returns an error", func() {
						_, err := cloud.Info()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Extracting api_version"))
					})
				})

				Context("when stemcell formats is not a []string", func() {
					BeforeEach(func() {
						infoResultWithApiV2 = map[string]interface{}{
							"stemcell_formats": "aws-raw",
							"api_version":      stemcellApiVersion,
						}
						fakeCPICmdRunner.RunReturns(CmdOutput{Result: infoResultWithApiV2}, nil)
					})

					It("returns an error", func() {
						_, err := cloud.Info()
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Extracting stemcell_formats"))
					})
				})
			})

			Context("when info method is not implemented in CPI", func() {
				BeforeEach(func() {
					fakeCPICmdRunner.RunReturns(CmdOutput{
						Result: nil,
						Error: &CmdError{
							Type:    "InvalidCall",
							Message: "Method is not known, got 'info'",
						},
					}, nil)
				})

				It("should return default APIVersion", func() {
					cpiInfo, err := cloud.Info()
					Expect(err).ToNot(HaveOccurred())
					Expect(cpiInfo.ApiVersion).To(Equal(cpiApiVersion))
				})
			})
		})
	})

	Describe("CreateStemcell", func() {
		var (
			stemcellImagePath string
			cloudProperties   biproperty.Map
		)

		BeforeEach(func() {
			stemcellImagePath = "/stemcell/path"
			cloudProperties = biproperty.Map{
				"fake-key": "fake-value",
			}

			fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{Result: infoResult}, nil)
		})

		Context("when the cpi successfully creates the stemcell", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunReturnsOnCall(1, CmdOutput{Result: "fake-cid"}, nil)
			})

			It("executes the cpi job script with stemcell image path & cloud_properties", func() {
				_, err := cloud.CreateStemcell(stemcellImagePath, cloudProperties)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.RunCallCount()).To(Equal(2))
				context, method, apiVersion, args := fakeCPICmdRunner.RunArgsForCall(1)
				Expect(context).To(Equal(expectedContext))
				Expect(method).To(Equal("create_stemcell"))
				Expect(apiVersion).To(Equal(1))
				Expect(args).To(Equal([]interface{}{stemcellImagePath, cloudProperties}))
			})

			It("returns the cid returned from executing the cpi script", func() {
				cid, err := cloud.CreateStemcell(stemcellImagePath, cloudProperties)
				Expect(err).NotTo(HaveOccurred())
				Expect(cid).To(Equal("fake-cid"))
			})
		})

		Context("when the result is of an unexpected type", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunReturnsOnCall(1, CmdOutput{Result: 1}, nil)
			})

			It("returns an error", func() {
				_, err := cloud.CreateStemcell(stemcellImagePath, cloudProperties)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unexpected external CPI command result: '1'"))
			})
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunReturnsOnCall(1, CmdOutput{}, errors.New("fake-run-error"))
			})

			It("returns an error", func() {
				_, err := cloud.CreateStemcell(stemcellImagePath, cloudProperties)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-error"))
			})
		})

		itHandlesCPIErrors("create_stemcell", func() error {
			_, err := cloud.CreateStemcell(stemcellImagePath, cloudProperties)
			return err
		})
	})

	Describe("DeleteStemcell", func() {
		BeforeEach(func() {
			fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{Result: infoResult}, nil)
		})

		It("executes the delete_stemcell method on the CPI with stemcell cid", func() {
			err := cloud.DeleteStemcell("fake-stemcell-cid")
			Expect(err).NotTo(HaveOccurred())

			context, method, apiVersion, args := fakeCPICmdRunner.RunArgsForCall(1)
			Expect(context).To(Equal(expectedContext))
			Expect(method).To(Equal("delete_stemcell"))
			Expect(apiVersion).To(Equal(1))
			Expect(args).To(Equal([]interface{}{"fake-stemcell-cid"}))
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{}, errors.New("fake-run-error"))
			})

			It("returns an error", func() {
				err := cloud.DeleteStemcell("fake-stemcell-cid")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-error"))
			})
		})

		itHandlesCPIErrors("delete_stemcell", func() error {
			return cloud.DeleteStemcell("fake-stemcell-cid")
		})
	})

	Describe("HasVM", func() {
		It("return true when VM exists", func() {
			fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{Result: infoResult}, nil)
			fakeCPICmdRunner.RunReturnsOnCall(1, CmdOutput{Result: true}, nil)

			found, err := cloud.HasVM("fake-vm-cid")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())

			context, method, apiVersion, args := fakeCPICmdRunner.RunArgsForCall(1)
			Expect(context).To(Equal(expectedContext))
			Expect(method).To(Equal("has_vm"))
			Expect(apiVersion).To(Equal(1))
			Expect(args).To(Equal([]interface{}{"fake-vm-cid"}))
		})

		It("return false when VM does not exist", func() {
			fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{Result: infoResult}, nil)
			fakeCPICmdRunner.RunReturnsOnCall(1, CmdOutput{Result: false}, nil)

			found, err := cloud.HasVM("fake-vm-cid")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{Result: infoResult}, nil)
				fakeCPICmdRunner.RunReturnsOnCall(1, CmdOutput{}, errors.New("fake-run-error"))
			})

			It("returns an error when executing the CPI command fails", func() {
				_, err := cloud.HasVM("fake-vm-cid")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-error"))
			})
		})

		itHandlesCPIErrors("has_vm", func() error {
			_, err := cloud.HasVM("fake-vm-cid")
			return err
		})
	})

	Describe("CreateVM", func() {
		var (
			agentID           string
			stemcellCID       string
			cloudProperties   biproperty.Map
			networkInterfaces map[string]biproperty.Map
			diskCIDs          []string
			env               biproperty.Map
		)

		BeforeEach(func() {
			agentID = "fake-agent-id"
			stemcellCID = "fake-stemcell-cid"
			networkInterfaces = map[string]biproperty.Map{
				"bosh": {
					"type": "dynamic",
					"cloud_properties": biproperty.Map{
						"a": "b",
					},
				},
			}
			diskCIDs = []string{"fake-disk-cid"}
			cloudProperties = biproperty.Map{
				"fake-cloud-property-key": "fake-cloud-property-value",
			}
			env = biproperty.Map{
				"fake-env-key": "fake-env-value",
			}
		})

		Context("when the cpi successfully creates the vm", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{Result: infoResult}, nil)
				fakeCPICmdRunner.RunReturnsOnCall(1, CmdOutput{Result: "fake-vm-cid"}, nil)
			})

			It("executes the cpi job script with the director UUID and stemcell CID", func() {
				_, err := cloud.CreateVM(agentID, stemcellCID, cloudProperties, diskCIDs, networkInterfaces, env)
				Expect(err).NotTo(HaveOccurred())

				context, method, apiVersion, args := fakeCPICmdRunner.RunArgsForCall(1)
				Expect(context).To(Equal(expectedContext))
				Expect(method).To(Equal("create_vm"))
				Expect(apiVersion).To(Equal(1))
				Expect(args).To(Equal([]interface{}{
					agentID,
					stemcellCID,
					cloudProperties,
					networkInterfaces,
					diskCIDs,
					env,
				}))
			})

			It("returns the cid returned from executing the cpi script", func() {
				cid, err := cloud.CreateVM(agentID, stemcellCID, cloudProperties, diskCIDs, networkInterfaces, env)
				Expect(err).NotTo(HaveOccurred())
				Expect(cid).To(Equal("fake-vm-cid"))
			})

			Context("when stemcell api_version is 2 and cpi api_version is 2", func() {
				BeforeEach(func() {
					var networks interface{}

					fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{Result: infoResultWithApiV2}, nil)
					fakeCPICmdRunner.RunReturnsOnCall(1, CmdOutput{Result: []interface{}{"fake-vm-cid", networks}}, nil)
					stemcellApiVersion = 2
				})

				It("returns the vm cid", func() {
					cid, err := cloud.CreateVM(agentID, stemcellCID, cloudProperties, diskCIDs, networkInterfaces, env)
					Expect(err).NotTo(HaveOccurred())
					Expect(cid).To(Equal("fake-vm-cid"))
				})

				Context("when the cpi's response is unexpected", func() {
					BeforeEach(func() {
						var networkHash = "can be anything, not checked right now"
						fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{Result: infoResultWithApiV2}, nil)
						// result: [vm-cid, network-hash{}]
						fakeCPICmdRunner.RunReturnsOnCall(1, CmdOutput{Result: []interface{}{1, networkHash}}, nil)
					})

					It("returns error", func() {
						_, err := cloud.CreateVM(agentID, stemcellCID, cloudProperties, diskCIDs, networkInterfaces, env)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Unexpected external CPI command result: '[]interface {}"))
					})
				})
			})
		})

		Context("when the result is of an unexpected type", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{Result: infoResult}, nil)
				fakeCPICmdRunner.RunReturnsOnCall(1, CmdOutput{Result: 1}, nil)
			})

			It("returns an error", func() {
				_, err := cloud.CreateVM(agentID, stemcellCID, cloudProperties, diskCIDs, networkInterfaces, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unexpected external CPI command result: '1'"))
			})
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunReturns(CmdOutput{}, errors.New("fake-run-error"))
			})

			It("returns an error", func() {
				_, err := cloud.CreateVM(agentID, stemcellCID, cloudProperties, diskCIDs, networkInterfaces, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-error"))
			})
		})

		itHandlesCPIErrors("create_vm", func() error {
			_, err := cloud.CreateVM(agentID, stemcellCID, cloudProperties, diskCIDs, networkInterfaces, env)
			return err
		})

	})

	Describe("SetDiskMetadata", func() {
		var metadata DiskMetadata
		BeforeEach(func() {
			fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{Result: infoResult}, nil)

			metadata = DiskMetadata{
				"director":       "bosh-init",
				"deployment":     "some-deployment",
				"instance_group": "some-instance_group",
				"instance_index": "0",
				"attached_at":    "2017-03-22T10:17:04Z",
			}
		})

		It("calls the set_disk_metadata CPI method", func() {
			diskCID := "fake-disk-cid"
			err := cloud.SetDiskMetadata(diskCID, metadata)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeCPICmdRunner.RunCallCount()).To(Equal(2))
			context, method, apiVersion, args := fakeCPICmdRunner.RunArgsForCall(1)
			Expect(context).To(Equal(expectedContext))
			Expect(method).To(Equal("set_disk_metadata"))
			Expect(apiVersion).To(Equal(1))
			Expect(args).To(Equal([]interface{}{diskCID, metadata}))

			//Expect(cloudfakes.CurrentRunInput).To(HaveLen(2))
			//Expect(cloudfakes.CurrentRunInput[1]).To(Equal(cloudfakes.RunInput{
			//	Context: expectedContext,
			//	Method:  "set_disk_metadata",
			//	Arguments: []interface{}{
			//		diskCID,
			//		metadata,
			//	},
			//	ApiVersion: 1,
			//}))
		})

		It("returns the error if running fails", func() {
			fakeCPICmdRunner.RunReturns(CmdOutput{}, errors.New("fake-run-error"))
			diskCID := "fake-disk-cid"
			err := cloud.SetDiskMetadata(diskCID, metadata)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("fake-run-error"))
		})

		itHandlesCPIErrors("set_disk_metadata", func() error {
			diskCID := "fake-disk-cid"
			return cloud.SetDiskMetadata(diskCID, metadata)
		})
	})

	Describe("SetVMMetadata", func() {
		BeforeEach(func() {
			fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{Result: infoResult}, nil)
		})

		It("calls the set_vm_metadata CPI method", func() {
			vmCID := "fake-vm-cid"
			metadata := VMMetadata{
				"director":   "bosh-init",
				"deployment": "some-deployment",
				"job":        "some-job",
				"name":       "some-job/0",
				"index":      "0",
			}
			err := cloud.SetVMMetadata(vmCID, metadata)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeCPICmdRunner.RunCallCount()).To(Equal(2))
			context, method, apiVersion, args := fakeCPICmdRunner.RunArgsForCall(1)
			Expect(context).To(Equal(expectedContext))
			Expect(method).To(Equal("set_vm_metadata"))
			Expect(apiVersion).To(Equal(1))
			Expect(args).To(Equal([]interface{}{vmCID, metadata}))
		})

		It("returns the error if running fails", func() {
			fakeCPICmdRunner.RunReturns(CmdOutput{}, errors.New("fake-run-error"))
			vmCID := "fake-vm-cid"
			metadata := VMMetadata{
				"director":   "bosh-init",
				"deployment": "some-deployment",
				"job":        "some-job",
				"name":       "some-job/0",
				"index":      "0",
			}

			err := cloud.SetVMMetadata(vmCID, metadata)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("fake-run-error"))
		})

		itHandlesCPIErrors("set_vm_metadata", func() error {
			vmCID := "fake-vm-cid"
			metadata := VMMetadata{
				"director":   "bosh-init",
				"deployment": "some-deployment",
				"job":        "some-job",
				"index":      "0",
			}
			return cloud.SetVMMetadata(vmCID, metadata)
		})
	})

	Describe("CreateDisk", func() {
		var (
			size            int
			cloudProperties biproperty.Map
			instanceID      string
		)

		BeforeEach(func() {
			size = 1024
			cloudProperties = biproperty.Map{
				"fake-cloud-property-key": "fake-cloud-property-value",
			}
			instanceID = "fake-instance-id"
		})

		Context("when the cpi successfully creates the disk", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{Result: infoResultWithApiV2}, nil)
				fakeCPICmdRunner.RunReturnsOnCall(1, CmdOutput{Result: "fake-disk-cid"}, nil)
			})

			It("executes the cpi job script with the correct arguments", func() {
				_, err := cloud.CreateDisk(size, cloudProperties, instanceID)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.RunCallCount()).To(Equal(2))
				context, method, apiVersion, args := fakeCPICmdRunner.RunArgsForCall(1)
				Expect(context).To(Equal(expectedContext))
				Expect(method).To(Equal("create_disk"))
				Expect(apiVersion).To(Equal(2))
				Expect(args).To(Equal([]interface{}{size, cloudProperties, instanceID}))
			})

			It("returns the cid returned from executing the cpi script", func() {
				cid, err := cloud.CreateDisk(size, cloudProperties, instanceID)
				Expect(err).NotTo(HaveOccurred())
				Expect(cid).To(Equal("fake-disk-cid"))
			})
		})

		Context("when the result is of an unexpected type", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{Result: infoResult}, nil)
				fakeCPICmdRunner.RunReturnsOnCall(1, CmdOutput{Result: 1}, nil)
			})

			It("returns an error", func() {
				_, err := cloud.CreateDisk(size, cloudProperties, instanceID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unexpected external CPI command result: '1'"))
			})
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunReturns(CmdOutput{}, errors.New("fake-run-error"))
			})

			It("returns an error", func() {
				_, err := cloud.CreateDisk(size, cloudProperties, instanceID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-error"))
			})
		})

		itHandlesCPIErrors("create_disk", func() error {
			_, err := cloud.CreateDisk(size, cloudProperties, instanceID)
			return err
		})
	})

	Describe("AttachDisk", func() {
		Context("when stemcell api version and cpi api version are 2", func() {
			DescribeTable("parsing disk hints as different types",
				func(inputHint interface{}, expected interface{}) {
					fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{Result: infoResultWithApiV2}, nil)
					fakeCPICmdRunner.RunReturnsOnCall(1, CmdOutput{Result: inputHint}, nil)
					stemcellApiVersion = 2

					diskHint, err := cloud.AttachDisk("fake-vm-cid", "fake-disk-cid")
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeCPICmdRunner.RunCallCount()).To(Equal(2))
					context, method, apiVersion, args := fakeCPICmdRunner.RunArgsForCall(1)
					Expect(context).To(Equal(expectedContext))
					Expect(method).To(Equal("attach_disk"))
					Expect(apiVersion).To(Equal(2))
					Expect(args).To(Equal([]interface{}{"fake-vm-cid", "fake-disk-cid"}))
					Expect(diskHint).To(Equal(expected))
				},
				Entry("string", "/dev/sdf", "/dev/sdf"),
				Entry("map", map[string]interface{}{
					"path": "/dev/1337",
					"lun":  "1",
				}, map[string]interface{}{
					"path": "/dev/1337",
					"lun":  "1",
				}),
			)
		})

		Context("when the cpi successfully attaches the disk", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{Result: infoResult}, nil)
			})

			It("executes the cpi job script with the correct arguments", func() {
				_, err := cloud.AttachDisk("fake-vm-cid", "fake-disk-cid")
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.RunCallCount()).To(Equal(2))
				context, method, apiVersion, args := fakeCPICmdRunner.RunArgsForCall(1)
				Expect(context).To(Equal(expectedContext))
				Expect(method).To(Equal("attach_disk"))
				Expect(apiVersion).To(Equal(1))
				Expect(args).To(Equal([]interface{}{"fake-vm-cid", "fake-disk-cid"}))
			})
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunReturns(CmdOutput{}, errors.New("fake-run-error"))
			})

			It("returns an error", func() {
				_, err := cloud.AttachDisk("fake-vm-cid", "fake-disk-cid")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-error"))
			})
		})

		itHandlesCPIErrors("attach_disk", func() error {
			_, err := cloud.AttachDisk("fake-vm-cid", "fake-disk-cid")
			return err
		})
	})

	Describe("DetachDisk", func() {
		Context("when the cpi successfully detaches the disk", func() {
			It("executes the cpi job script with the correct arguments", func() {

				fakeCPICmdRunner.RunReturns(CmdOutput{Result: infoResultWithApiV2}, nil)

				err := cloud.DetachDisk("fake-vm-cid", "fake-disk-cid")
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.RunCallCount()).To(Equal(2))
				context, method, apiVersion, args := fakeCPICmdRunner.RunArgsForCall(1)
				Expect(context).To(Equal(expectedContext))
				Expect(method).To(Equal("detach_disk"))
				Expect(apiVersion).To(Equal(2))
				Expect(args).To(Equal([]interface{}{"fake-vm-cid", "fake-disk-cid"}))
			})
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunReturns(CmdOutput{}, errors.New("fake-run-error"))
			})

			It("returns an error", func() {
				err := cloud.DetachDisk("fake-vm-cid", "fake-disk-cid")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-error"))
			})
		})

		itHandlesCPIErrors("detach_disk", func() error {
			return cloud.DetachDisk("fake-vm-cid", "fake-disk-cid")
		})
	})

	Describe("DeleteVM", func() {
		Context("when the cpi successfully deletes vm", func() {
			It("executes the cpi job script with the correct arguments", func() {

				fakeCPICmdRunner.RunReturns(CmdOutput{Result: infoResultWithApiV2}, nil)

				err := cloud.DeleteVM("fake-vm-cid")
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.RunCallCount()).To(Equal(2))
				context, method, apiVersion, args := fakeCPICmdRunner.RunArgsForCall(1)
				Expect(context).To(Equal(expectedContext))
				Expect(method).To(Equal("delete_vm"))
				Expect(apiVersion).To(Equal(2))
				Expect(args).To(Equal([]interface{}{"fake-vm-cid"}))
			})
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunReturns(CmdOutput{}, errors.New("fake-run-error"))
			})

			It("returns an error", func() {
				err := cloud.DeleteVM("fake-vm-cid")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-error"))
			})
		})

		itHandlesCPIErrors("delete_vm", func() error {
			return cloud.DeleteVM("fake-vm-cid")
		})
	})

	Describe("DeleteDisk", func() {
		Context("when the cpi successfully deletes disk", func() {
			It("executes the cpi job script with the correct arguments", func() {

				fakeCPICmdRunner.RunReturns(CmdOutput{Result: infoResultWithApiV2}, nil)

				err := cloud.DeleteDisk("fake-disk-cid")
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.RunCallCount()).To(Equal(2))
				context, method, apiVersion, args := fakeCPICmdRunner.RunArgsForCall(1)
				Expect(context).To(Equal(expectedContext))
				Expect(method).To(Equal("delete_disk"))
				Expect(apiVersion).To(Equal(2))
				Expect(args).To(Equal([]interface{}{"fake-disk-cid"}))
			})
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunReturnsOnCall(0, CmdOutput{Result: infoResult}, nil)
				fakeCPICmdRunner.RunReturnsOnCall(1, CmdOutput{}, errors.New("fake-run-error"))
			})

			It("returns an error", func() {
				err := cloud.DeleteDisk("fake-disk-cid")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-error"))
			})
		})

		itHandlesCPIErrors("delete_disk", func() error {
			return cloud.DeleteDisk("fake-disk-cid")
		})
	})
})
