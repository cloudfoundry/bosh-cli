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

	Describe("DiffConfigByIDOrContent", func() {
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
					ghttp.VerifyBody([]byte(`{"from":{"id":"1","content":""},"to":{"id":"2","content":""}}`)),
					ghttp.RespondWith(http.StatusOK, `{"diff":[["release:",null],["  version: 0.0.1","removed"],["  version: 0.0.2","added"]]}`),
				),
			)

			diff, err := director.DiffConfigByIDOrContent("1", nil, "2", nil)
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

				_, err := director.DiffConfigByIDOrContent("1", nil, "2", nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(
					"Config with ID '2' not found."))
			})
		})

		Context("when one of the IDs or more is not an integer", func() {
			It("returns an error", func() {
				_, err := director.DiffConfigByIDOrContent("1", nil, "a", nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(
					"--to-id needs to be an integer."))

				_, err = director.DiffConfigByIDOrContent("b", nil, "2", nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(
					"--from-id needs to be an integer."))
			})
		})

		It("returns error if response in non-200", func() {
			AppendBadRequest(ghttp.VerifyRequest("POST", "/configs/diff"), server)

			_, err := director.DiffConfigByIDOrContent("1", nil, "2", nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Fetching diff result: Director responded with non-successful status code"))
		})

		It("returns error if from-id and from-content are specified", func() {
			_, err := director.DiffConfigByIDOrContent("1", []byte("config"), "2", nil)
			Expect(err.Error()).To(Equal("only one of --from-id and --from-content can be specified"))
		})

		It("returns error if to-id and to-content are specified", func() {
			_, err := director.DiffConfigByIDOrContent("1", nil, "2", []byte("config"))
			Expect(err.Error()).To(Equal("only one of --to-id and --to-content can be specified"))
		})

		It("returns no error if only from-content is specified", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/configs/diff"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"application/json"},
					}),
					ghttp.VerifyBody([]byte(`{"from":{"id":"","content":"config"},"to":{"id":"1","content":""}}`)),
					ghttp.RespondWith(http.StatusOK, `{"diff":[["release:",null],["  version: 0.0.1","removed"],["  version: 0.0.2","added"]]}`),
				),
			)
			_, err := director.DiffConfigByIDOrContent("", []byte("config"), "1", nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns no error if only from-content is specified", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/configs/diff"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"application/json"},
					}),
					ghttp.VerifyBody([]byte(`{"from":{"id":"1","content":""},"to":{"id":"","content":"config"}}`)),
					ghttp.RespondWith(http.StatusOK, `{"diff":[["release:",null],["  version: 0.0.1","removed"],["  version: 0.0.2","added"]]}`),
				),
			)
			_, err := director.DiffConfigByIDOrContent("1", nil, "", []byte("config"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if neither from-id nor from-content are specified", func() {
			_, err := director.DiffConfigByIDOrContent("", nil, "", []byte("config"))
			Expect(err.Error()).To(Equal("one of --from-id or --from-content must be specified"))
		})

		It("returns error if neither to-id nor to-content are specified", func() {
			_, err := director.DiffConfigByIDOrContent("", []byte("config"), "", nil)
			Expect(err.Error()).To(Equal("one of --to-id or --to-content must be specified"))
		})
	})
})
