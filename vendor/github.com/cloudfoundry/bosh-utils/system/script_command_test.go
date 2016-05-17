package system_test

import (
	. "github.com/cloudfoundry/bosh-utils/system"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ScriptCommand", func() {
	Context("when using windows", func() {
		It("wraps command in powershell", func() {
			scriptCommandFactory := NewScriptCommandFactory("windows")
			cmd := scriptCommandFactory.New("some_script.ps1", "--some-option")
			Expect(cmd).To(Equal(Command{
				Name: "powershell",
				Args: []string{"-noprofile", "-noninteractive", "some_script.ps1", "--some-option"},
			}))
		})
	})

	Context("when using linux", func() {
		It("returns command as is", func() {
			scriptCommandFactory := NewScriptCommandFactory("linux")
			cmd := scriptCommandFactory.New("run.sh", "--some-option")
			Expect(cmd).To(Equal(Command{
				Name: "run.sh",
				Args: []string{"--some-option"},
			}))
		})
	})
})
