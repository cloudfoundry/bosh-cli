package system_test

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/cloudfoundry/bosh-utils/system"
)

var _ = Describe("ConcreteScriptRunner", func() {

	Describe("ScriptRunner", func() {

		var runner ScriptRunner
		BeforeEach(func() {
			logger := boshlog.NewLogger(boshlog.LevelNone)
			runner = NewConcreteScriptRunner(
				NewExecCmdRunner(logger),
				NewOsFileSystem(logger),
				logger,
			)
		})

		It("runs a valid script", func() {
			var scriptBody string
			if Windows {
				scriptBody = "Write-Output stdout\r\n[Console]::Error.WriteLine('stderr')"
			} else {
				scriptBody = "#!/bin/bash\n(>&1 echo 'stdout')\n(>&2 echo 'stderr')"
			}
			stdout, stderr, err := runner.Run(scriptBody)
			Expect(err).ToNot(HaveOccurred())
			Expect(strings.TrimSpace(stdout)).To(Equal("stdout"))
			Expect(strings.TrimSpace(stderr)).To(Equal("stderr"))
		})

		It("runs an invalid script and returns an error", func() {
			scriptBody := "Blah-Blah-I-Dont-Exist 'stdout'"
			_, _, err := runner.Run(scriptBody)
			Expect(err).To(HaveOccurred())
		})

		It("runs a multi-line script with variable assignement", func() {
			var scriptBody string
			if Windows {
				scriptBody = `
$MY_STDOUT=@"
stdout
"@
$MY_STDERR="stderr"
try {
	Write-Output "${MY_STDOUT}"
	[Console]::Error.WriteLine("${MY_STDERR}")
} catch {
	$Host.UI.WriteErrorLine($_.Exception.Message)
	Exit 1
}
Exit 0`
			} else {
				scriptBody = `
#!/usr/bin/env bash

MY_STD=\
"std"
export MY_STDOUT="${MY_STD}out"
export MY_STDERR="${MY_STD}err"
echo "$MY_STDOUT"
(>&2 echo \
"$MY_STDERR")`
			}

			stdout, stderr, err := runner.Run(scriptBody)
			Expect(err).ToNot(HaveOccurred())
			Expect(strings.TrimSpace(stdout)).To(Equal("stdout"))
			Expect(strings.TrimSpace(stderr)).To(Equal("stderr"))
		})

		It("returns an error when there is non-zero exit code", func() {
			var scriptBody string
			if Windows {
				scriptBody = "Start-Sleep -Milliseconds 250\r\n[Console]::Error.WriteLine('stderr')\r\nExit 1"
			} else {
				scriptBody = "#!/bin/bash\n(>&2 echo 'stderr')\nexit 1"
			}
			_, stderr, err := runner.Run(scriptBody)
			Expect(err).To(HaveOccurred())
			Expect(strings.TrimSpace(stderr)).To(Equal("stderr"))
		})
	})
})
