package director_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	. "github.com/cloudfoundry/bosh-cli/v7/director"
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

	Describe("LatestCloudConfig", func() {
		It("returns latest cloud config if there is at least one", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/cloud_configs", "name=&limit=1"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[
	{"properties": "first"},
	{"properties": "second"}
]`),
				),
			)

			cc, err := director.LatestCloudConfig("")
			Expect(err).ToNot(HaveOccurred())
			Expect(cc).To(Equal(CloudConfig{Properties: "first"}))
		})

		It("returns named cloud config if there is at least one and name is specified", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/cloud_configs", "name=foo-name&limit=1"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[
	{"properties": "first"},
	{"properties": "second"}
]`),
				),
			)

			cc, err := director.LatestCloudConfig("foo-name")
			Expect(err).ToNot(HaveOccurred())
			Expect(cc).To(Equal(CloudConfig{Properties: "first"}))
		})

		It("returns error for when name cannot be found", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/cloud_configs", "name=foo-name&limit=1"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[]`),
				),
			)

			_, err := director.LatestCloudConfig("foo-name")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("No cloud config"))
		})

		It("returns error if there is no default cloud config", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/cloud_configs", "name=&limit=1"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[]`),
				),
			)

			_, err := director.LatestCloudConfig("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("No cloud config"))
		})

		It("returns error if info response in non-200", func() {
			AppendBadRequest(ghttp.VerifyRequest("GET", "/cloud_configs"), server)

			_, err := director.LatestCloudConfig("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Finding cloud configs: Director responded with non-successful status code"))
		})

		It("returns error if info cannot be unmarshalled", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/cloud_configs"),
					ghttp.RespondWith(http.StatusOK, ``),
				),
			)

			_, err := director.LatestCloudConfig("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Finding cloud configs: Unmarshaling Director response"))
		})
	})

	Describe("UpdateCloudConfig", func() {
		It("updates cloud config", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/cloud_configs", "name=smurf-runtime-config"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"text/yaml"},
					}),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := director.UpdateCloudConfig("smurf-runtime-config", []byte("config"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if info response in non-200", func() {
			AppendBadRequest(ghttp.VerifyRequest("POST", "/cloud_configs", "name="), server)

			err := director.UpdateCloudConfig("", nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Updating cloud config: Director responded with non-successful status code"))
		})
	})

	Describe("DiffCloudConfig", func() {
		expectedDiffResponse := ConfigDiff{
			Diff: [][]interface{}{
				[]interface{}{"azs:", nil},
				[]interface{}{"- name: az2", "removed"},
				[]interface{}{"  cloud_properties: {}", "removed"},
			},
		}

		It("diffs cloud config with the given name", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/cloud_configs/diff", "name=cc1"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"text/yaml"},
					}),
					ghttp.RespondWith(http.StatusOK, `{"diff":[["azs:",null],["- name: az2","removed"],["  cloud_properties: {}","removed"]]}`),
				),
			)

			diff, err := director.DiffCloudConfig("cc1", []byte("config"))
			Expect(err).ToNot(HaveOccurred())
			Expect(diff).To(Equal(expectedDiffResponse))
		})

		It("returns error if info response in non-200", func() {
			AppendBadRequest(ghttp.VerifyRequest("POST", "/cloud_configs/diff"), server)

			_, err := director.DiffCloudConfig("smurf", nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Fetching diff result: Director responded with non-successful status code"))
		})

		It("is backwards compatible with directors without the `/diff` endpoint", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/cloud_configs/diff", "name=cc1"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"text/yaml"},
					}),
					ghttp.RespondWith(http.StatusNotFound, ""),
				),
			)

			diff, err := director.DiffCloudConfig("cc1", []byte("config"))
			Expect(err).ToNot(HaveOccurred())
			Expect(diff).To(Equal(ConfigDiff{}))
		})
	})
})
