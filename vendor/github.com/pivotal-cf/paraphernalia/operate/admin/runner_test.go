package admin_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pivotal-cf/paraphernalia/operate/admin"
	"github.com/tedsuo/ifrit"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Runner", func() {
	var (
		runner  ifrit.Runner
		process ifrit.Process

		port        string
		optionFuncs []admin.OptionFunc
	)

	BeforeEach(func() {
		port = strconv.Itoa(60061 + GinkgoParallelNode())
	})

	JustBeforeEach(func() {
		runner = admin.Runner(port, optionFuncs...)
		process = ifrit.Invoke(runner)
	})

	AfterEach(func() {
		process.Signal(os.Interrupt)

		err := <-process.Wait()
		Expect(err).NotTo(HaveOccurred())
	})

	It("starts the debug server", func() {
		response, err := http.Get("http://localhost:" + port + "/debug/pprof/cmdline")
		Expect(err).NotTo(HaveOccurred())

		body, err := ioutil.ReadAll(response.Body)
		Expect(err).NotTo(HaveOccurred())

		Expect(body).To(ContainSubstring("admin"))
	})

	Describe("enabling the information endpoint", func() {
		BeforeEach(func() {
			optionFuncs = []admin.OptionFunc{
				admin.WithInfo(admin.ServiceInfo{
					Name:        "service-name",
					Description: "it's a thing which does a thing",
					Team:        "team name",
				}),
			}
		})

		It("let's the developers of a service tell people what it is", func() {
			response, err := http.Get("http://localhost:" + port + "/info")
			Expect(err).NotTo(HaveOccurred())

			body, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())

			Expect(body).To(ContainSubstring("service-name"))
			Expect(body).To(ContainSubstring("it's a thing which does a thing"))
			Expect(body).To(ContainSubstring("team name"))
		})
	})

	Describe("enabling the uptime endpoint", func() {
		BeforeEach(func() {
			optionFuncs = []admin.OptionFunc{
				admin.WithUptime(),
			}
		})

		It("let's the operators of a service known how long it has been running", func() {
			response, err := http.Get("http://localhost:" + port + "/uptime")
			Expect(err).NotTo(HaveOccurred())

			body, err := ioutil.ReadAll(response.Body)
			Expect(err).NotTo(HaveOccurred())

			parsed, err := time.ParseDuration(string(body))
			Expect(err).NotTo(HaveOccurred())

			Expect(parsed).To(BeNumerically("<", 1*time.Second))
		})
	})
})
