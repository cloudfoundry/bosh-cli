package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"io/ioutil"
	"os"
	"os/exec"
)

var _ = Describe("Verify_multidigest", func() {
	var session *gexec.Session
	var act func(arg ...string)
	var tempFile *os.File

	BeforeEach(func() {
		var err error
		tempFile, err = ioutil.TempFile("", "multi-digest-test")
		Expect(err).ToNot(HaveOccurred())

		act = func(argCommands ...string) {
			var err error
			command := exec.Command(pathToBoshUtils, argCommands...)
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		}
	})

	AfterEach(func() {
		os.Remove(tempFile.Name())
	})

	Describe("version option", func() {
		It("has a version flag", func() {
			act("--version")
			Eventually(session).Should(gexec.Exit(0))
			Eventually(session.Out).Should(gbytes.Say("version \\[DEV BUILD\\]"))
		})
	})

	Context("when correct args are passed to verify-multi-digest command", func() {
		It("exits 0", func() {
			act("verify-multi-digest", tempFile.Name(), "da39a3ee5e6b4b0d3255bfef95601890afd80709")
			Eventually(session).Should(gexec.Exit(0))
		})
	})

	Context("when passing incorrect args", func() {
		It("exits 1 when digest does not match", func() {
			act("verify-multi-digest", tempFile.Name(), "incorrect-digest")
			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("Expected stream to have digest 'incorrect-digest' but was 'da39a3ee5e6b4b0d3255bfef95601890afd80709'"))
		})

		It("exits 1 when file does not exist", func() {
			act("verify-multi-digest", "potato", "da39a3ee5e6b4b0d3255bfef95601890afd80709")
			Eventually(session).Should(gexec.Exit(1))
			Eventually(session.Err).Should(gbytes.Say("open potato:"))
		})
	})

})
