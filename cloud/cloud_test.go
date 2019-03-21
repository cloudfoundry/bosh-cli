package cloud_test

import (
	"errors"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	biproperty "github.com/cloudfoundry/bosh-utils/property"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakebicloud "github.com/cloudfoundry/bosh-cli/cloud/fakes"

	. "github.com/cloudfoundry/bosh-cli/cloud"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("Cloud", func() {
	var (
		cloud               Cloud
		expectedContext     CmdContext
		fakeCPICmdRunner    *fakebicloud.FakeCPICmdRunner
		logger              boshlog.Logger
		stemcellApiVersion  = 1
		cpiApiVersion       = 1
		infoResult          map[string]interface{}
		infoResultWithApiV2 map[string]interface{}
	)

	BeforeEach(func() {
		fakeCPICmdRunner = fakebicloud.NewFakeCPICmdRunner()
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

			fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
				{Result: infoResult},
			}

			fakeCPICmdRunner.RunCmdOutputs = append(
				fakeCPICmdRunner.RunCmdOutputs,
				CmdOutput{
					Error: &CmdError{
						Type:    "Bosh::Cloud::CloudError",
						Message: "fake-cpi-error-msg",
					},
				},
			)

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
				fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{{
					Result: infoResultWithApiV2,
				}}
				cpiInfo, err := cloud.Info()
				Expect(cpiInfo).To(Equal(infoParsed))
				Expect(err).ToNot(HaveOccurred())

				inputs := fakeCPICmdRunner.CurrentRunInput
				Expect(inputs).To(HaveLen(1))
				input := inputs[0]
				expectedInput := fakebicloud.RunInput{
					Context: expectedContext,
					Method:  "info",
					// The correct answer should be `[]interface{}{}` but because of https://github.com/golang/go/issues/4133 we have to use nil.
					Arguments:  nil,
					ApiVersion: 1,
				}
				Expect(input.ApiVersion).To(Equal(expectedInput.ApiVersion))
				Expect(input.Method).To(Equal(expectedInput.Method))
				Expect(input.Context).To(Equal(expectedInput.Context))
				Expect(input.Arguments).To(Equal(expectedInput.Arguments))
			})

			It("uses a default cpi api version if an old cpi does not have api version", func() {
				infoParsed := CpiInfo{
					StemcellFormats: []string{"aws-raw", "aws-light"},
					ApiVersion:      1,
				}
				fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{{
					Result: infoResult,
				}}
				cpiInfo, err := cloud.Info()
				Expect(cpiInfo).To(Equal(infoParsed))
				Expect(err).ToNot(HaveOccurred())
			})

			Context("when the cpi command execution fails", func() {
				BeforeEach(func() {
					fakeCPICmdRunner.RunErrs = []error{errors.New("info")}
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
					fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{{
						Result: infoResult,
					}}
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

						fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{{
							Result: infoResultWithApiV2,
						}}
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
						fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{{
							Result: infoResultWithApiV2,
						}}
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
					fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{{
						Result: nil,
						Error: &CmdError{
							Type:    "InvalidCall",
							Message: "Method is not known, got 'info'",
						},
					}}
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

			fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
				{Result: infoResult},
				{Result: 1},
			}
		})

		Context("when the cpi successfully creates the stemcell", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunCmdOutputs[1] = CmdOutput{Result: "fake-cid"}
			})

			It("executes the cpi job script with stemcell image path & cloud_properties", func() {
				_, err := cloud.CreateStemcell(stemcellImagePath, cloudProperties)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.CurrentRunInput).To(HaveLen(2))
				Expect(fakeCPICmdRunner.CurrentRunInput[1]).To(Equal(fakebicloud.RunInput{
					Context: expectedContext,
					Method:  "create_stemcell",
					Arguments: []interface{}{
						stemcellImagePath,
						cloudProperties,
					},
					ApiVersion: 1,
				}))
			})

			It("returns the cid returned from executing the cpi script", func() {
				cid, err := cloud.CreateStemcell(stemcellImagePath, cloudProperties)
				Expect(err).NotTo(HaveOccurred())
				Expect(cid).To(Equal("fake-cid"))
			})
		})

		Context("when the result is of an unexpected type", func() {
			It("returns an error", func() {
				_, err := cloud.CreateStemcell(stemcellImagePath, cloudProperties)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unexpected external CPI command result: '1'"))
			})
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunErrs = []error{nil, errors.New("fake-run-error")}
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
			fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
				{Result: infoResult},
			}
		})

		It("executes the delete_stemcell method on the CPI with stemcell cid", func() {
			err := cloud.DeleteStemcell("fake-stemcell-cid")
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeCPICmdRunner.CurrentRunInput).To(HaveLen(2))
			Expect(fakeCPICmdRunner.CurrentRunInput[1]).To(Equal(fakebicloud.RunInput{
				Context: expectedContext,
				Method:  "delete_stemcell",
				Arguments: []interface{}{
					"fake-stemcell-cid",
				},
				ApiVersion: 1,
			}))
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunErrs = []error{nil, errors.New("fake-run-error")}
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
			fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
				{Result: infoResult},
				{Result: true},
			}

			found, err := cloud.HasVM("fake-vm-cid")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeTrue())

			Expect(fakeCPICmdRunner.CurrentRunInput[1]).To(Equal(fakebicloud.RunInput{
				Context:    expectedContext,
				Method:     "has_vm",
				Arguments:  []interface{}{"fake-vm-cid"},
				ApiVersion: 1,
			}))
		})

		It("return false when VM does not exist", func() {
			fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
				{Result: infoResult},
				{Result: false},
			}

			found, err := cloud.HasVM("fake-vm-cid")
			Expect(err).ToNot(HaveOccurred())
			Expect(found).To(BeFalse())
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
					{Result: infoResult},
				}
				fakeCPICmdRunner.RunErrs = []error{nil, errors.New("fake-run-error")}
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
			cloudProperties = biproperty.Map{
				"fake-cloud-property-key": "fake-cloud-property-value",
			}
			env = biproperty.Map{
				"fake-env-key": "fake-env-value",
			}
		})

		Context("when the cpi successfully creates the vm", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
					{
						Result: infoResult,
					},
					{
						Result: "fake-vm-cid",
					},
				}
			})

			It("executes the cpi job script with the director UUID and stemcell CID", func() {
				_, err := cloud.CreateVM(agentID, stemcellCID, cloudProperties, networkInterfaces, env)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.CurrentRunInput).To(HaveLen(2))
				Expect(fakeCPICmdRunner.CurrentRunInput[1]).To(Equal(fakebicloud.RunInput{
					Context: expectedContext,
					Method:  "create_vm",
					Arguments: []interface{}{
						agentID,
						stemcellCID,
						cloudProperties,
						networkInterfaces,
						[]interface{}{},
						env,
					},
					ApiVersion: 1,
				}))
			})

			It("returns the cid returned from executing the cpi script", func() {
				cid, err := cloud.CreateVM(agentID, stemcellCID, cloudProperties, networkInterfaces, env)
				Expect(err).NotTo(HaveOccurred())
				Expect(cid).To(Equal("fake-vm-cid"))
			})

			Context("when stemcell api_version is 2 and cpi api_version is 2", func() {
				BeforeEach(func() {
					var networks interface{}

					fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
						{
							Result: infoResultWithApiV2,
						},
						{
							Result: []interface{}{"fake-vm-cid", networks},
						},
					}
					stemcellApiVersion = 2
				})

				It("returns the vm cid", func() {
					cid, err := cloud.CreateVM(agentID, stemcellCID, cloudProperties, networkInterfaces, env)
					Expect(err).NotTo(HaveOccurred())
					Expect(cid).To(Equal("fake-vm-cid"))
				})

				Context("when the cpi's response is unexpected", func() {
					BeforeEach(func() {
						var networkHash = "can be anything, not checked right now"
						fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
							{
								Result: infoResultWithApiV2,
							},
							{
								// result: [vm-cid, network-hash{}]
								Result: []interface{}{1, networkHash},
							},
						}
					})

					It("returns error", func() {
						_, err := cloud.CreateVM(agentID, stemcellCID, cloudProperties, networkInterfaces, env)
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("Unexpected external CPI command result: '[]interface {}"))
					})
				})
			})
		})

		Context("when the result is of an unexpected type", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
					{
						Result: infoResult,
					},
					{
						Result: 1,
					},
				}
			})

			It("returns an error", func() {
				_, err := cloud.CreateVM(agentID, stemcellCID, cloudProperties, networkInterfaces, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unexpected external CPI command result: '1'"))
			})
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunErrs = []error{errors.New("fake-run-error")}
			})

			It("returns an error", func() {
				_, err := cloud.CreateVM(agentID, stemcellCID, cloudProperties, networkInterfaces, env)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-error"))
			})
		})

		itHandlesCPIErrors("create_vm", func() error {
			_, err := cloud.CreateVM(agentID, stemcellCID, cloudProperties, networkInterfaces, env)
			return err
		})

	})

	Describe("SetDiskMetadata", func() {
		BeforeEach(func() {
			fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
				{Result: infoResult},
			}
		})
		metadata := DiskMetadata{
			"director":       "bosh-init",
			"deployment":     "some-deployment",
			"instance_group": "some-instance_group",
			"instance_index": "0",
			"attached_at":    "2017-03-22T10:17:04Z",
		}
		It("calls the set_disk_metadata CPI method", func() {
			diskCID := "fake-disk-cid"
			err := cloud.SetDiskMetadata(diskCID, metadata)
			Expect(err).ToNot(HaveOccurred())

			Expect(fakeCPICmdRunner.CurrentRunInput).To(HaveLen(2))
			Expect(fakeCPICmdRunner.CurrentRunInput[1]).To(Equal(fakebicloud.RunInput{
				Context: expectedContext,
				Method:  "set_disk_metadata",
				Arguments: []interface{}{
					diskCID,
					metadata,
				},
				ApiVersion: 1,
			}))
		})

		It("returns the error if running fails", func() {
			fakeCPICmdRunner.RunErrs = []error{errors.New("fake-run-error")}
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
			fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
				{Result: infoResult},
			}
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

			Expect(fakeCPICmdRunner.CurrentRunInput).To(HaveLen(2))
			Expect(fakeCPICmdRunner.CurrentRunInput[1]).To(Equal(fakebicloud.RunInput{
				Context: expectedContext,
				Method:  "set_vm_metadata",
				Arguments: []interface{}{
					vmCID,
					metadata,
				},
				ApiVersion: 1,
			}))
		})

		It("returns the error if running fails", func() {
			fakeCPICmdRunner.RunErrs = []error{errors.New("fake-run-error")}
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
				fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
					{Result: infoResultWithApiV2},
					{Result: "fake-disk-cid"},
				}
			})

			It("executes the cpi job script with the correct arguments", func() {
				_, err := cloud.CreateDisk(size, cloudProperties, instanceID)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.CurrentRunInput).To(HaveLen(2))
				Expect(fakeCPICmdRunner.CurrentRunInput[1]).To(Equal(fakebicloud.RunInput{
					Context: expectedContext,
					Method:  "create_disk",
					Arguments: []interface{}{
						size,
						cloudProperties,
						instanceID,
					},
					ApiVersion: 2,
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
				fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
					{Result: infoResult},
					{Result: 1},
				}
			})

			It("returns an error", func() {
				_, err := cloud.CreateDisk(size, cloudProperties, instanceID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Unexpected external CPI command result: '1'"))
			})
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunErrs = []error{errors.New("fake-run-error")}
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
					fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
						{Result: infoResultWithApiV2},
						{Result: inputHint},
					}
					stemcellApiVersion = 2

					diskHint, err := cloud.AttachDisk("fake-vm-cid", "fake-disk-cid")
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeCPICmdRunner.CurrentRunInput).To(HaveLen(2))
					Expect(fakeCPICmdRunner.CurrentRunInput[1]).To(Equal(fakebicloud.RunInput{
						Context: expectedContext,
						Method:  "attach_disk",
						Arguments: []interface{}{
							"fake-vm-cid",
							"fake-disk-cid",
						},
						ApiVersion: 2,
					}))
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
				fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
					{Result: infoResult},
				}
			})

			It("executes the cpi job script with the correct arguments", func() {
				_, err := cloud.AttachDisk("fake-vm-cid", "fake-disk-cid")
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.CurrentRunInput).To(HaveLen(2))
				Expect(fakeCPICmdRunner.CurrentRunInput[1]).To(Equal(fakebicloud.RunInput{
					Context: expectedContext,
					Method:  "attach_disk",
					Arguments: []interface{}{
						"fake-vm-cid",
						"fake-disk-cid",
					},
					ApiVersion: 1,
				}))
			})
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunErrs = []error{errors.New("fake-run-error")}
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

				fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
					{Result: infoResultWithApiV2},
				}

				err := cloud.DetachDisk("fake-vm-cid", "fake-disk-cid")
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.CurrentRunInput).To(HaveLen(2))
				Expect(fakeCPICmdRunner.CurrentRunInput[1]).To(Equal(fakebicloud.RunInput{
					Context: expectedContext,
					Method:  "detach_disk",
					Arguments: []interface{}{
						"fake-vm-cid",
						"fake-disk-cid",
					},
					ApiVersion: 2,
				}))
			})
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunErrs = []error{errors.New("fake-run-error")}
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

				fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
					{Result: infoResultWithApiV2},
				}

				err := cloud.DeleteVM("fake-vm-cid")
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.CurrentRunInput).To(HaveLen(2))
				Expect(fakeCPICmdRunner.CurrentRunInput[1]).To(Equal(fakebicloud.RunInput{
					Context: expectedContext,
					Method:  "delete_vm",
					Arguments: []interface{}{
						"fake-vm-cid",
					},
					ApiVersion: 2,
				}))
			})
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunErrs = []error{errors.New("fake-run-error")}
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

				fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{
					{Result: infoResultWithApiV2},
				}

				err := cloud.DeleteDisk("fake-disk-cid")
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeCPICmdRunner.CurrentRunInput).To(HaveLen(2))
				Expect(fakeCPICmdRunner.CurrentRunInput[1]).To(Equal(fakebicloud.RunInput{
					Context: expectedContext,
					Method:  "delete_disk",
					Arguments: []interface{}{
						"fake-disk-cid",
					},
					ApiVersion: 2,
				}))
			})
		})

		Context("when the cpi command execution fails", func() {
			BeforeEach(func() {
				fakeCPICmdRunner.RunCmdOutputs = []CmdOutput{{Result: infoResult}}
				fakeCPICmdRunner.RunErrs = []error{nil, errors.New("fake-run-error")}
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
