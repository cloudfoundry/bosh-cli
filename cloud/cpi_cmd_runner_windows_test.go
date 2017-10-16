package cloud_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"

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
		})

		Context("when obtaining working directory fails", func() {
			BeforeEach(func() {
				result := fakesys.FakeCmdResult{
					Error: errors.New("fake-error-trying-to-obtain-working-directory"),
				}
				cmdRunner.AddCmdResult("bash -c pwd", result)
			})

			It("returns an error", func() {
				_, err := cpiCmdRunner.Run(context, "fake-method", "fake-argument")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-error-trying-to-obtain-working-directory"))
			})
		})

	})
})
