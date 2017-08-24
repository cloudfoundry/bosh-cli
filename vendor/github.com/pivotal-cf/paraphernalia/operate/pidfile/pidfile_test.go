package pidfile_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tedsuo/ifrit"

	"github.com/pivotal-cf/paraphernalia/operate/pidfile"
)

var _ = Describe("Pidfile", func() {
	var (
		tmpdir      string
		pidfilePath string

		runner ifrit.Runner

		process ifrit.Process
	)

	BeforeEach(func() {
		var err error

		tmpdir, err = ioutil.TempDir("", "pidfile-test")
		Ω(err).ShouldNot(HaveOccurred())

		pidfilePath = filepath.Join(tmpdir, "pidfile")
		runner = pidfile.NewRunner(pidfilePath)
	})

	AfterEach(func() {
		err := os.RemoveAll(tmpdir)
		Ω(err).ShouldNot(HaveOccurred())
	})

	JustBeforeEach(func() {
		process = ifrit.Envoke(runner)
	})

	var currentPid = fmt.Sprintf("%d", os.Getpid())

	itWritesTheCurrentPid := func() {
		It("writes the current pid to the file", func() {
			pidfile, err := os.Open(pidfilePath)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(ioutil.ReadAll(pidfile)).Should(Equal([]byte(currentPid)))
		})
	}

	Context("when the pidfile does not exist", func() {
		itWritesTheCurrentPid()
	})

	Context("when the parent directory of the pidfile does not exist", func() {
		BeforeEach(func() {
			err := os.RemoveAll(tmpdir)
			Ω(err).ShouldNot(HaveOccurred())
		})

		itWritesTheCurrentPid()
	})

	Context("when the pidfile already exists", func() {
		Context("and it contains an active pid", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(pidfilePath, []byte(currentPid), 0644)
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("fails with a useful error message", func() {
				err := <-process.Wait()

				Expect(err).To(MatchError(ContainSubstring(pidfilePath)))
				Expect(err).To(MatchError(ContainSubstring(currentPid)))
			})
		})

		Context("and it contains an inactive pid", func() {
			itWritesTheCurrentPid()
		})

		Context("and it is empty", func() {
			BeforeEach(func() {
				file, err := os.Create(pidfilePath)
				Ω(err).ShouldNot(HaveOccurred())

				err = file.Close()
				Ω(err).ShouldNot(HaveOccurred())
			})

			itWritesTheCurrentPid()
		})
	})

	Describe("stopping", func() {
		JustBeforeEach(func() {
			process.Signal(os.Interrupt)
			Eventually(process.Wait()).Should(Receive(BeNil()))
		})

		It("removes the pidfile", func() {
			_, err := os.Stat(pidfilePath)
			Ω(err).Should(HaveOccurred())
		})
	})
})
