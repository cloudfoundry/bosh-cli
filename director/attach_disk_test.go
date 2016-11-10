package director_test

import (
	. "github.com/cloudfoundry/bosh-cli/director"

	"github.com/cloudfoundry/bosh-cli/director/directorfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("AttachDisk", func() {
	var (
		director   Director
		server     *ghttp.Server
		deployment *directorfakes.FakeDeployment
	)

	BeforeEach(func() {
		director, server = BuildServer()
		deployment = &directorfakes.FakeDeployment{}
		deployment.NameReturns("foo")
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("AttachDisk", func() {
		It("calls attachdisk director api", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/disks/disk_cid/attachments", "deployment=foo&job=dea&instance_id=17f01a35-bf9c-4949-bcf2-c07a95e4df33"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(200, "{}"),
				),
			)

			err := director.AttachDisk(deployment, NewInstanceSlug("dea", "17f01a35-bf9c-4949-bcf2-c07a95e4df33"), "disk_cid")
			Expect(server.ReceivedRequests()).To(HaveLen(1))
			Expect(err).ToNot(HaveOccurred())
		})

		Context("director returns a non-200 response", func() {

			It("should return an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PUT", "/disks/disk_cid/attachments", "deployment=foo&job=dea&instance_id=17f01a35-bf9c-4949-bcf2-c07a95e4df33"),
						ghttp.VerifyBasicAuth("username", "password"),
						ghttp.RespondWith(500, "Internal Server Error"),
					),
				)

				err := director.AttachDisk(deployment, NewInstanceSlug("dea", "17f01a35-bf9c-4949-bcf2-c07a95e4df33"), "disk_cid")
				Expect(server.ReceivedRequests()).To(HaveLen(1))
				Expect(err).To(HaveOccurred())
			})
		})
	})

})
