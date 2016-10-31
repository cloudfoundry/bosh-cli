package director_test

import (
	"net/http"

	. "github.com/cloudfoundry/bosh-cli/director"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Director", func() {
	var (
		director   Director
		deployment Deployment
		server     *ghttp.Server
	)

	BeforeEach(func() {
		director, server = BuildServer()

		var err error

		deployment, err = director.FindDeployment("dep")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("DeleteVm", func() {
		It("deletes vm", func() {
			ConfigureTaskResult(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/vms/cid"),
					ghttp.VerifyBasicAuth("username", "password"),
				),
				"",
				server,
			)
			err := deployment.DeleteVm("cid")
			Expect(err).ToNot(HaveOccurred())
		})

		It("succeeds even if error occurrs if vm no longer exists", func() {
			AppendBadRequest(ghttp.VerifyRequest("DELETE", "/vms/cid"), server)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/vms"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[]`),
				),
			)

			err := deployment.DeleteVm("cid")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns delete error if listing vms fails", func() {
			AppendBadRequest(ghttp.VerifyRequest("DELETE", "/vms/cid"), server)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/vms"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, ``),
				),
			)

			err := deployment.DeleteVm("cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Deleting vm 'cid'"))
		})

		It("returns delete error if response is non-200 and vm still exists", func() {
			AppendBadRequest(ghttp.VerifyRequest("DELETE", "/vms/cid"), server)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/vms"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[{"vm_cid": "cid"}]`),
				),
			)

			err := deployment.DeleteVm("cid")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Deleting vm 'cid'"))
		})
	})
})
