package cloud_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
	"runtime"

	. "github.com/cloudfoundry/bosh-cli/cloud"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CpiCmdRunner", func() {
	var (
		cpiCmdRunner CPICmdRunner
		context      CmdContext
		cmdRunner    *fakesys.FakeCmdRunner
		cpi          CPI
	)

	BeforeEach(func() {
		context = CmdContext{
			DirectorID: "fake-director-id",
		}

		cpi = CPI{
			JobPath:     filepath.Join("/", "jobs", "cpi"),
			JobsDir:     filepath.Join("/", "jobs"),
			PackagesDir: filepath.Join("/", "packages"),
		}

		cmdRunner = fakesys.NewFakeCmdRunner()
		logger := boshlog.NewLogger(boshlog.LevelNone)
		cpiCmdRunner = NewCPICmdRunner(cmdRunner, cpi, logger)
	})

	Describe("Run", func() {
		It("creates correct command", func() {
			cpiCmdOutput := CmdOutput{}
			cpiOutputBytes, err := json.Marshal(cpiCmdOutput)
			Expect(err).NotTo(HaveOccurred())

			cpiResult := fakesys.FakeCmdResult{
				Stdout:     string(cpiOutputBytes),
				ExitStatus: 0,
			}

			if runtime.GOOS == "windows" {
				cmdRunner.AddCmdResult("bash -x /working/directory/jobs/cpi/bin/cpi", cpiResult)

				pwdResult := fakesys.FakeCmdResult{
					Stdout:     "/working/directory",
					ExitStatus: 0,
				}
				cmdRunner.AddCmdResult("bash -c pwd", pwdResult)

				_, err = cpiCmdRunner.Run(context, "fake-method", "fake-argument-1", "fake-argument-2")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmdRunner.RunComplexCommands).To(HaveLen(2))

				pwdActualCmd := cmdRunner.RunComplexCommands[0]
				Expect(pwdActualCmd.Name).To(Equal("bash"))
				Expect(pwdActualCmd.Args).To(Equal([]string{"-c", "pwd"}))
				Expect(pwdActualCmd.Env).To(BeNil())
				Expect(pwdActualCmd.UseIsolatedEnv).To(BeFalse())

				cpiActualCmd := cmdRunner.RunComplexCommands[1]
				Expect(cpiActualCmd.Name).To(Equal("bash"))
				Expect(cpiActualCmd.Args).To(Equal([]string{"-x", "/working/directory/jobs/cpi/bin/cpi"}))
				Expect(cpiActualCmd.Env).To(Equal(map[string]string{
					"BOSH_PACKAGES_DIR": "/working/directory" + filepath.ToSlash(cpi.PackagesDir),
					"BOSH_JOBS_DIR":     "/working/directory" + filepath.ToSlash(cpi.JobsDir),
				}))
				Expect(cpiActualCmd.UseIsolatedEnv).To(BeFalse())
				bytes, err := ioutil.ReadAll(cpiActualCmd.Stdin)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(bytes)).To(Equal(
					`{` +
						`"method":"fake-method",` +
						`"arguments":["fake-argument-1","fake-argument-2"],` +
						`"context":{"director_uuid":"fake-director-id"}` +
						`}`,
				))
			} else {
				cmdRunner.AddCmdResult(filepath.Join("/", "jobs", "cpi", "bin", "cpi"), cpiResult)

				_, err = cpiCmdRunner.Run(context, "fake-method", "fake-argument-1", "fake-argument-2")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmdRunner.RunComplexCommands).To(HaveLen(1))

				actualCmd := cmdRunner.RunComplexCommands[0]
				Expect(actualCmd.Name).To(Equal(filepath.Join("/", "jobs", "cpi", "bin", "cpi")))
				Expect(actualCmd.Args).To(BeNil())
				Expect(actualCmd.Env).To(Equal(map[string]string{
					"BOSH_PACKAGES_DIR": filepath.ToSlash(cpi.PackagesDir),
					"BOSH_JOBS_DIR":     filepath.ToSlash(cpi.JobsDir),
					"PATH":              "/usr/local/bin:/usr/bin:/bin:/sbin",
				}))
				Expect(actualCmd.UseIsolatedEnv).To(BeTrue())
				bytes, err := ioutil.ReadAll(actualCmd.Stdin)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(bytes)).To(Equal(
					`{` +
						`"method":"fake-method",` +
						`"arguments":["fake-argument-1","fake-argument-2"],` +
						`"context":{"director_uuid":"fake-director-id"}` +
						`}`,
				))
			}
		})

		Context("when the command succeeds", func() {
			BeforeEach(func() {
				cpiCmdOutput := CmdOutput{
					Result: "fake-cid",
				}
				cpiOutputBytes, err := json.Marshal(cpiCmdOutput)
				Expect(err).NotTo(HaveOccurred())

				cpiResult := fakesys.FakeCmdResult{
					Stdout:     string(cpiOutputBytes),
					ExitStatus: 0,
				}

				if runtime.GOOS == "windows" {
					cmdRunner.AddCmdResult("bash -x /working/directory/jobs/cpi/bin/cpi", cpiResult)

					pwdResult := fakesys.FakeCmdResult{
						Stdout:     "/working/directory",
						ExitStatus: 0,
					}
					cmdRunner.AddCmdResult("bash -c pwd", pwdResult)
				} else {
					cmdRunner.AddCmdResult(filepath.Join("/", "jobs", "cpi", "bin", "cpi"), cpiResult)
				}
			})

			It("returns the result", func() {
				cmdOutput, err := cpiCmdRunner.Run(context, "fake-method", "fake-argument")
				Expect(err).NotTo(HaveOccurred())
				Expect(cmdOutput).To(Equal(CmdOutput{
					Result: "fake-cid",
					Error:  nil,
					Log:    "",
				}))
			})
		})

		Context("when running the command fails", func() {
			BeforeEach(func() {
				cpiResult := fakesys.FakeCmdResult{
					Error: errors.New("fake-error-trying-to-run-command"),
				}

				if runtime.GOOS == "windows" {
					cmdRunner.AddCmdResult("bash -x /working/directory/jobs/cpi/bin/cpi", cpiResult)

					pwdResult := fakesys.FakeCmdResult{
						Stdout:     "/working/directory",
						ExitStatus: 0,
					}
					cmdRunner.AddCmdResult("bash -c pwd", pwdResult)
				} else {
					cmdRunner.AddCmdResult(filepath.Join("/", "jobs", "cpi", "bin", "cpi"), cpiResult)
				}
			})

			It("returns an error", func() {
				_, err := cpiCmdRunner.Run(context, "fake-method", "fake-argument")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-error-trying-to-run-command"))
			})
		})

		Context("when the command runs but fails", func() {
			BeforeEach(func() {
				cpiCmdOutput := CmdOutput{
					Error: &CmdError{
						Message: "fake-run-error",
					},
					Result: "fake-cid",
				}
				cpiOutputBytes, err := json.Marshal(cpiCmdOutput)
				Expect(err).NotTo(HaveOccurred())

				cpiResult := fakesys.FakeCmdResult{
					Stdout:     string(cpiOutputBytes),
					ExitStatus: 0,
				}

				if runtime.GOOS == "windows" {
					cmdRunner.AddCmdResult("bash -x /working/directory/jobs/cpi/bin/cpi", cpiResult)

					pwdResult := fakesys.FakeCmdResult{
						Stdout:     "/working/directory",
						ExitStatus: 0,
					}
					cmdRunner.AddCmdResult("bash -c pwd", pwdResult)
				} else {
					cmdRunner.AddCmdResult(filepath.Join("/", "jobs", "cpi", "bin", "cpi"), cpiResult)
				}
			})

			It("returns the command output and no error", func() {
				cmdOutput, err := cpiCmdRunner.Run(context, "fake-method", "fake-argument")
				Expect(err).ToNot(HaveOccurred())
				Expect(cmdOutput.Error.Message).To(ContainSubstring("fake-run-error"))
			})
		})
	})
})
