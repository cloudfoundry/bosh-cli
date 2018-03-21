package director_test

import (
	"net/http"

	. "github.com/cloudfoundry/bosh-cli/director"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/ghttp"
)

var _ bool = Describe("Director", func() {
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

	Describe("DiffConfigByID", func() {
		expectedDiffResponse := ConfigDiff{
			Diff: [][]interface{}{
				{"release:", nil},
				{"  version: 0.0.1", "removed"},
				{"  version: 0.0.2", "added"},
			},
		}

		It("diffs the given configs", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/configs/diff"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"application/json"},
					}),
					ghttp.VerifyBody([]byte(`{"from":{"id":"1"},"to":{"id":"2"}}`)),
					ghttp.RespondWith(http.StatusOK, `{"diff":[["release:",null],["  version: 0.0.1","removed"],["  version: 0.0.2","added"]]}`),
				),
			)

			diff, err := director.DiffConfigByID("1", "2")
			Expect(err).ToNot(HaveOccurred())
			Expect(diff).To(Equal(expectedDiffResponse))
		})

		Context("when director returns 440012 Config not found error", func() {
			It("returns the director error description", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/configs/diff"),
						ghttp.RespondWith(http.StatusNotFound, `{"code":440012,"description":"Config with ID '2' not found."}`),
					),
				)

				_, err := director.DiffConfigByID("1", "2")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(
					"Config with ID '2' not found."))
			})
		})

		Context("when one of the IDs or more is not an integer", func() {
			It("returns an error", func() {
				_, err := director.DiffConfigByID("1", "a")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(
					"Config ID needs to be an integer."))

				_, err = director.DiffConfigByID("b", "2")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(
					"Config ID needs to be an integer."))
			})
		})

		It("returns error if response in non-200", func() {
			AppendBadRequest(ghttp.VerifyRequest("POST", "/configs/diff"), server)

			_, err := director.DiffConfigByID("1", "2")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Fetching diff result: Director responded with non-successful status code"))
		})

	})

})
