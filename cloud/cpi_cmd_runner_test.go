package cloud_test

import (
	"encoding/json"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	. "github.com/cloudfoundry/bosh-micro-cli/cloud"
)

var _ = Describe("CpiCmdRunner", func() {
	var (
		cpiCmdRunner   CPICmdRunner
		cmdRunner      *fakesys.FakeCmdRunner
		deploymentUUID string
		cpiJob         CPIJob
		cmdInputBytes  []byte
	)

	BeforeEach(func() {
		deploymentUUID = "fake-deployment-uuid"
		cmdInput := CmdInput{
			Method: "fake-method",
			Arguments: []interface{}{
				"fake-argument-1",
				"fake-argument-2",
			},
			Context: CmdContext{
				DirectorUUID: deploymentUUID,
			},
		}

		var err error
		cmdInputBytes, err = json.Marshal(cmdInput)
		Expect(err).NotTo(HaveOccurred())

		cpiJob = CPIJob{
			JobPath:     "/jobs/cpi",
			JobsDir:     "/jobs",
			PackagesDir: "/packages",
		}

		cmdRunner = fakesys.NewFakeCmdRunner()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		cpiCmdRunner = NewCPICmdRunner(cmdRunner, cpiJob, deploymentUUID, logger)
	})

	Describe("Run", func() {
		It("creates correct command", func() {
			cmdOutput := CmdOutput{}
			outputBytes, err := json.Marshal(cmdOutput)
			Expect(err).NotTo(HaveOccurred())

			result := fakesys.FakeCmdResult{
				Stdout:     string(outputBytes),
				ExitStatus: 0,
			}
			cmdRunner.AddCmdResult("/jobs/cpi/bin/cpi", result)

			_, err = cpiCmdRunner.Run("fake-method", "fake-argument-1", "fake-argument-2")
			Expect(err).NotTo(HaveOccurred())
			Expect(cmdRunner.RunComplexCommands).To(HaveLen(1))

			actualCmd := cmdRunner.RunComplexCommands[0]
			Expect(actualCmd.Name).To(Equal("/jobs/cpi/bin/cpi"))
			Expect(actualCmd.Args).To(BeNil())
			Expect(actualCmd.Env).To(Equal(map[string]string{
				"BOSH_PACKAGES_DIR": cpiJob.PackagesDir,
				"BOSH_JOBS_DIR":     cpiJob.JobsDir,
				"PATH":              "/usr/local/bin:/usr/bin:/bin",
			}))
			Expect(actualCmd.UseIsolatedEnv).To(BeTrue())
			bytes, err := ioutil.ReadAll(actualCmd.Stdin)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(bytes)).To(Equal(string(cmdInputBytes)))
		})

		Context("when the command succeeds", func() {
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

			It("returns the result", func() {
				cmdOutput, err := cpiCmdRunner.Run("fake-method", "fake-argument")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmdOutput).To(Equal(CmdOutput{
					Result: "fake-cid",
					Error:  nil,
					Log:    "",
				}))
			})
		})

		Context("when the command fails", func() {
			BeforeEach(func() {
				cmdOutput := CmdOutput{
					Error: &CmdError{
						Message: "fake-run-error",
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
				_, err := cpiCmdRunner.Run("fake-method", "fake-argument")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-run-error"))
			})
		})
	})
})
