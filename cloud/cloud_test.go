package cloud_test

import (
	"encoding/json"
	"os"
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
		cpiJobPath        string
		stemcell          bmstemcell.Stemcell
		stemcellImagePath string
		deploymentUUID    string
		cloudProperties   map[string]interface{}
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		cmdRunner = fakesys.NewFakeCmdRunner()
		cpiJobPath = "/jobs/cpi"
		deploymentUUID = "fake-uuid"
		logger := boshlog.NewLogger(boshlog.LevelNone)
		cloud = NewCloud(fs, cmdRunner, cpiJobPath, deploymentUUID, logger)
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
			cmdInputBytes, err := json.Marshal(cmdInput)
			Expect(err).NotTo(HaveOccurred())

			cmdInputString = string(cmdInputBytes)

			cmdOutput := CmdOutput{
				Result: map[string]string{"cid": "fake-cid"},
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

		It("makes the the cpi job script executable", func() {
			_, err := cloud.CreateStemcell(stemcell)
			Expect(err).NotTo(HaveOccurred())

			Expect(fs.GetFileTestStat("/jobs/cpi/bin/cpi").FileMode).To(Equal(os.FileMode(0770)))
		})

		It("executes the cpi job script with stemcell image path & cloud_properties", func() {
			_, err := cloud.CreateStemcell(stemcell)
			Expect(err).NotTo(HaveOccurred())
			Expect(cmdRunner.RunCommandsWithInput).To(HaveLen(1))
			Expect(cmdRunner.RunCommandsWithInput[0]).To(Equal(
				[]string{
					cmdInputString,
					"/jobs/cpi/bin/cpi",
				},
			))
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
				cmdInputBytes, err := json.Marshal(cmdInput)
				Expect(err).NotTo(HaveOccurred())

				cmdInputString = string(cmdInputBytes)

				cmdOutput := CmdOutput{
					Result: map[string]string{"cid": "fake-cid"},
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
				Expect(cmdRunner.RunCommandsWithInput).To(HaveLen(1))
				Expect(cmdRunner.RunCommandsWithInput[0]).To(Equal(
					[]string{
						cmdInputString,
						"/jobs/cpi/bin/cpi",
					},
				))
			})
		})
	})
})
