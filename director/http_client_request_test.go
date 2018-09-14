package director_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	. "github.com/cloudfoundry/bosh-cli/director"
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

	Describe("NewHTTPClientRequest", func() {
		type testInterface interface {
			NewHTTPClientRequest() ClientRequest
		}

		It("does not add it to the interface as it's particular to Director HTTP client", func() {
			var director Director

			if _, ok := director.(testInterface); ok {
				Fail("Director interface should not have NewHTTPClientRequest()")
			}
		})

		It("returns usable HTTP client", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/info"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `resp-body`),
				),
			)

			body, resp, err := director.(DirectorImpl).NewHTTPClientRequest().RawGet("/info", nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).ToNot(BeNil())
			Expect(body).To(Equal([]byte(`resp-body`)))
		})
	})
})
