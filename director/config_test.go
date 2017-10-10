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
		director Director
		server   *ghttp.Server
	)

	BeforeEach(func() {
		director, server = BuildServer()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("LatestConfig", func() {
		It("returns the latest config", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/configs", "type=my-type&name=my-name&latest=true"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[{"content": "first"}]`),
				),
			)

			cc, err := director.LatestConfig("my-type", "my-name")
			Expect(err).ToNot(HaveOccurred())
			Expect(cc).To(Equal(Config{Content: "first"}))
		})

		Context("when there is no config", func() {
			It("returns error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/configs", "type=missing-type&latest=true&name=default"),
						ghttp.VerifyBasicAuth("username", "password"),
						ghttp.RespondWith(http.StatusOK, `[]`),
					),
				)

				_, err := director.LatestConfig("missing-type", "default")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("No config"))
			})
		})

		Context("when server returns an error", func() {
			It("returns error", func() {
				AppendBadRequest(ghttp.VerifyRequest("GET", "/configs", "type=fake-type&latest=true&name=default"), server)

				_, err := director.LatestConfig("fake-type", "default")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(
					"Finding config: Director responded with non-successful status code"))
			})
		})
	})

	Describe("ListConfigs", func() {
		Context("when no filters are given", func() {
			It("uses no query params and returns list of config items", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/configs", "latest=false"),
						ghttp.VerifyBasicAuth("username", "password"),
						ghttp.RespondWith(http.StatusOK, `[{"name": "first", "type": "my-type"}]`),
					),
				)

				cc, err := director.ListConfigs(ConfigsFilter{})
				Expect(err).ToNot(HaveOccurred())
				Expect(cc).To(Equal([]ConfigListItem{ConfigListItem{Type: "my-type", Name: "first"}}))
			})
		})

		Context("when filters are given", func() {
			It("uses them as query parameters and returns list of config items", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/configs", "latest=false&name=first&type=my-type"),
						ghttp.VerifyBasicAuth("username", "password"),
						ghttp.RespondWith(http.StatusOK, `[{"name": "first", "type": "my-type"}]`),
					),
				)

				cc, err := director.ListConfigs(ConfigsFilter{Type: "my-type", Name: "first"})
				Expect(err).ToNot(HaveOccurred())
				Expect(cc).To(Equal([]ConfigListItem{ConfigListItem{Type: "my-type", Name: "first"}}))
			})
		})

		Context("when server returns an error", func() {
			It("returns error", func() {
				AppendBadRequest(ghttp.VerifyRequest("GET", "/configs"), server)

				_, err := director.ListConfigs(ConfigsFilter{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(
					"Listing configs: Director responded with non-successful status code '400'"))
			})
		})
	})

	Describe("UpdateConfig", func() {
		It("updates config", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/configs", "type=my-type&name=my-name"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"text/yaml"},
					}),
					ghttp.RespondWith(http.StatusNoContent, `{'id': '1', 'type': 'my-type', 'name': 'my-name', 'content': '---'}`),
				),
			)

			err := director.UpdateConfig("my-type", "my-name", []byte("---"))
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when server returns an error", func() {
			It("returns error", func() {
				AppendBadRequest(ghttp.VerifyRequest("POST", "/configs"), server)

				err := director.UpdateConfig("fake-type", "fake-name", nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(
					"Updating config: Director responded with non-successful status code '400'"))
			})
		})
	})

	Describe("DeleteConfig", func() {
		Context("when config exists in director", func() {
			It("returns true", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/configs", "type=my-type&name=my-name"),
						ghttp.VerifyBasicAuth("username", "password"),
						ghttp.RespondWith(http.StatusCreated, nil),
					),
				)

				deleted, err := director.DeleteConfig("my-type", "my-name")
				Expect(err).ToNot(HaveOccurred())
				Expect(deleted).To(Equal(true))
			})
		})

		Context("when no matching config exists in director", func() {
			It("returns false", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", "/configs", "type=my-type&name=my-name"),
						ghttp.VerifyBasicAuth("username", "password"),
						ghttp.RespondWith(http.StatusNotFound, nil),
					),
				)

				deleted, err := director.DeleteConfig("my-type", "my-name")
				Expect(err).ToNot(HaveOccurred())
				Expect(deleted).To(Equal(false))
			})
		})

		Context("when server returns an error", func() {
			It("returns error", func() {
				AppendBadRequest(ghttp.VerifyRequest("DELETE", "/configs"), server)

				_, err := director.DeleteConfig("my-type", "my-name")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(
					"Deleting config: Director responded with non-successful status code '400'"))
			})
		})
	})
})
