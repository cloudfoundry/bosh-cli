package cloud_test

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/cloud"
)

var _ = Describe("Cloud", func() {
	var (
		cloud             Cloud
		fs                *fakesys.FakeFileSystem
		cmdRunner         *fakesys.FakeCmdRunner
		stemcell          bmstemcell.Stemcell
		stemcellImagePath string
		deploymentUUID    string
		cpiJob            CPIJob
		cloudProperties   map[string]interface{}
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
		stemcellImagePath = "/stemcell/path"
		cloudProperties = map[string]interface{}{
			"fake-key": "fake-value",
		}

		stemcell = bmstemcell.Stemcell{
			ImagePath:       stemcellImagePath,
			CloudProperties: cloudProperties,
		}
	})

	Describe("CreateStemcell", func() {
		var (
			cmdInputString string
			cmdInputBytes  []byte
		)

		BeforeEach(func() {
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

			cmdInputString = string(cmdInputBytes)

			cmdOutput := CmdOutput{
				Result: "fake-cid",
				Log:    "",
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
			Expect(bytes).To(Equal(cmdInputBytes))
		})

		It("returns the cid returned from executing the cpi script", func() {
			cid, err := cloud.CreateStemcell(stemcell)
			Expect(err).NotTo(HaveOccurred())
			Expect(cid).To(Equal(bmstemcell.CID("fake-cid")))
		})

		Context("when cloud_propeties are complex", func() {
			BeforeEach(func() {
				cloudProperties = map[string]interface{}{
					"fake-map-key": map[string]string{
						"fake-inner-key": "fake-inner-value",
					},
					"fake-array-key": []string{
						"fake-array-element",
					},
				}

				stemcell = bmstemcell.Stemcell{
					ImagePath:       stemcellImagePath,
					CloudProperties: cloudProperties,
				}

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
				var err error
				cmdInputBytes, err = json.Marshal(cmdInput)
				Expect(err).NotTo(HaveOccurred())

				cmdInputString = string(cmdInputBytes)

				cmdOutput := CmdOutput{
					Result: "fake-cid",
					Log:    "",
				}
				outputBytes, err := json.Marshal(cmdOutput)
				Expect(err).NotTo(HaveOccurred())

				result := fakesys.FakeCmdResult{
					Stdout:     string(outputBytes),
					ExitStatus: 0,
				}
				cmdString := strings.Join([]string{cmdInputString, "/jobs/cpi/bin/cpi"}, " ")
				cmdRunner.AddCmdResult(cmdString, result)
			})

			It("marshalls complex cloud_properties correctly", func() {
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
				Expect(bytes).To(Equal(cmdInputBytes))
			})
		})
	})
})
