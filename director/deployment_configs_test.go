package director_test

import (
	"net/http"

	boshdir "github.com/cloudfoundry/bosh-cli/director"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("DeploymentConfigs", func() {
	var (
		dir    boshdir.Director
		server *ghttp.Server
	)
	BeforeEach(func() {
		dir, server = BuildServer()
	})

	AfterEach(func() {
		server.Close()
	})
	Describe("ListDeploymentConfigs", func() {
		Context("When director supports deployment_configs endpoint", func() {
			It("Returns a List of Deployment Configurations", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/deployment_configs", "deployment[]=fakeDeployment"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, []map[string]interface{}{
							{
								"config": map[string]interface{}{
									"id":   1,
									"type": "fakeType",
									"name": "fakeConfig",
								},
							},
							{
								"config": map[string]interface{}{
									"id":   2,
									"type": "fakeType2",
									"name": "fakeConfig2",
								},
							},
						}),
					),
				)

				configs, err := dir.ListDeploymentConfigs("fakeDeployment")
				Expect(err).NotTo(HaveOccurred())

				Expect(configs.GetConfig(0).Name).To(Equal("fakeConfig"))
				Expect(configs.GetConfig(0).Type).To(Equal("fakeType"))
				Expect(configs.GetConfig(0).Id).To(Equal(1))
				Expect(configs.GetConfig(1).Name).To(Equal("fakeConfig2"))
				Expect(configs.GetConfig(1).Type).To(Equal("fakeType2"))
				Expect(configs.GetConfig(1).Id).To(Equal(2))
			})
		})
		Context("When director does not support deployment_configs endpoint", func() {
			It("Returns an empty list of configs", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/deployment_configs", "deployment[]=fakeDeployment"),
						ghttp.RespondWithJSONEncoded(http.StatusNotFound, nil),
					),
				)
				configs, err := dir.ListDeploymentConfigs("fakeDeployment")
				Expect(err).NotTo(HaveOccurred())
				Expect(configs.Configs).To(BeNil())
			})
		})
	})
})
