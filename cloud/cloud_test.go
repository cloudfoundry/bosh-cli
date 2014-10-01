package cloud_test

import (
	"encoding/json"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"
	bmvm "github.com/cloudfoundry/bosh-micro-cli/vm"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/cloud"
)

var _ = Describe("Cloud", func() {
	var (
		cloud          Cloud
		fs             *fakesys.FakeFileSystem
		cmdRunner      *fakesys.FakeCmdRunner
		deploymentUUID string
		cpiJob         CPIJob
		cmdInputBytes  []byte
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		cmdRunner = fakesys.NewFakeCmdRunner()
		deploymentUUID = "fake-uuid"
		cpiJob = CPIJob{
			JobPath:      "/jobs/cpi",
			JobsPath:     "/jobs",
			PackagesPath: "/packages",
		}
		logger := boshlog.NewLogger(boshlog.LevelNone)
		cloud = NewCloud(fs, cmdRunner, cpiJob, deploymentUUID, logger)
	})

	Describe("CreateStemcell", func() {
		var (
			stemcell          bmstemcell.Stemcell
			stemcellImagePath string
			cloudProperties   map[string]interface{}
		)

		BeforeEach(func() {
			stemcellImagePath = "/stemcell/path"
			cloudProperties = map[string]interface{}{
				"fake-key": "fake-value",
			}
			stemcell = bmstemcell.Stemcell{
				ImagePath:       stemcellImagePath,
				CloudProperties: cloudProperties,
			}
			err := fs.WriteFile("/jobs/cpi/bin/cpi", []byte{})
			Expect(err).NotTo(HaveOccurred())

			cmdInput := CmdInput{
				Method: "create_stemcell",
				Arguments: []interface{}{
					stemcellImagePath,
					cloudProperties,
				},
				Context: CmdContext{
					DirectorUUID: deploymentUUID,
				},
			}
			cmdInputBytes, err = json.Marshal(cmdInput)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the cpi successfully creates the stemcell", func() {
			BeforeEach(func() {
				cmdOutput := CmdOutput{
					Result: "fake-cid",
				}
				outputBytes, err := json.Marshal(cmdOutput)
				Expect(err).NotTo(HaveOccurred())

				result := fakesys.FakeCmdResult{
					Stdout:     string(outputBytes),
					ExitStatus: 0,
				}
				cmdRunner.AddCmdResult("/jobs/cpi/bin/cpi", result)
			})

			It("executes the cpi job script with stemcell image path & cloud_properties", func() {
				_, err := cloud.CreateStemcell(stemcell)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmdRunner.RunComplexCommands).To(HaveLen(1))

				actualCmd := cmdRunner.RunComplexCommands[0]
				Expect(actualCmd.Name).To(Equal("/jobs/cpi/bin/cpi"))
				Expect(actualCmd.Args).To(BeNil())
				Expect(actualCmd.Env).To(Equal(map[string]string{
					"BOSH_PACKAGES_DIR": cpiJob.PackagesPath,
					"BOSH_JOBS_DIR":     cpiJob.JobsPath,
				}))

				bytes, err := ioutil.ReadAll(actualCmd.Stdin)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(bytes)).To(Equal(string(cmdInputBytes)))
			})

			It("returns the cid returned from executing the cpi script", func() {
				cid, err := cloud.CreateStemcell(stemcell)
				Expect(err).NotTo(HaveOccurred())
				Expect(cid).To(Equal(bmstemcell.CID("fake-cid")))
			})
		})

		Context("when the cpi returns an error", func() {
			BeforeEach(func() {
				cmdOutput := CmdOutput{
					Error: &CmdError{
						Message: "fake-error",
					},
					Result: "fake-cid",
				}
				outputBytes, err := json.Marshal(cmdOutput)
				Expect(err).NotTo(HaveOccurred())

				result := fakesys.FakeCmdResult{
					Stdout:     string(outputBytes),
					ExitStatus: 0,
				}
				cmdRunner.AddCmdResult("/jobs/cpi/bin/cpi", result)
			})

			It("returns an error", func() {
				_, err := cloud.CreateStemcell(stemcell)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("External CPI command for method `create_stemcell' returned an error"))
			})
		})
	})

	Describe("CreateVM", func() {
		var (
			stemcellCID bmstemcell.CID
		)

		BeforeEach(func() {
			stemcellCID = "fake-stemcell-cid"
			err := fs.WriteFile("/jobs/cpi/bin/cpi", []byte{})
			Expect(err).NotTo(HaveOccurred())

			cmdInput := CmdInput{
				Method: "create_vm",
				Arguments: []interface{}{
					stemcellCID,
				},
				Context: CmdContext{
					DirectorUUID: deploymentUUID,
				},
			}
			cmdInputBytes, err = json.Marshal(cmdInput)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the cpi successfully creates the vm", func() {
			BeforeEach(func() {
				cmdOutput := CmdOutput{
					Result: "fake-vm-cid",
				}
				outputBytes, err := json.Marshal(cmdOutput)
				Expect(err).NotTo(HaveOccurred())

				result := fakesys.FakeCmdResult{
					Stdout:     string(outputBytes),
					ExitStatus: 0,
				}
				cmdRunner.AddCmdResult("/jobs/cpi/bin/cpi", result)
			})

			It("executes the cpi job script with the stemcell CID", func() {
				_, err := cloud.CreateVM(stemcellCID)
				Expect(err).NotTo(HaveOccurred())
				Expect(cmdRunner.RunComplexCommands).To(HaveLen(1))

				actualCmd := cmdRunner.RunComplexCommands[0]
				Expect(actualCmd.Name).To(Equal("/jobs/cpi/bin/cpi"))
				Expect(actualCmd.Args).To(BeNil())
				Expect(actualCmd.Env).To(Equal(map[string]string{
					"BOSH_PACKAGES_DIR": cpiJob.PackagesPath,
					"BOSH_JOBS_DIR":     cpiJob.JobsPath,
				}))

				bytes, err := ioutil.ReadAll(actualCmd.Stdin)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(bytes)).To(Equal(string(cmdInputBytes)))
			})

			It("returns the cid returned from executing the cpi script", func() {
				cid, err := cloud.CreateVM(stemcellCID)
				Expect(err).NotTo(HaveOccurred())
				Expect(cid).To(Equal(bmvm.CID("fake-vm-cid")))
			})
		})

		Context("when the cpi returns an error", func() {
			BeforeEach(func() {
				cmdOutput := CmdOutput{
					Error: &CmdError{
						Message: "fake-error",
					},
				}
				outputBytes, err := json.Marshal(cmdOutput)
				Expect(err).NotTo(HaveOccurred())

				result := fakesys.FakeCmdResult{
					Stdout:     string(outputBytes),
					ExitStatus: 0,
				}
				cmdRunner.AddCmdResult("/jobs/cpi/bin/cpi", result)
			})

			It("returns an error", func() {
				_, err := cloud.CreateVM(stemcellCID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("External CPI command for method `create_vm' returned an error"))
			})
		})
	})
})
