package configserver_test

import (
	"net/http"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	. "github.com/cloudfoundry/bosh-cli/configserver"
)

var _ = Describe("HTTPClient", func() {
	var (
		logger boshlog.Logger
	)

	BeforeEach(func() {
		logger = boshlog.NewLogger(boshlog.LevelNone)
	})

	Describe("generic client", func() {
		It("returns error if url is empty", func() {
			_, err := NewHTTPClient(HTTPClientOpts{}, logger)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected config server URL to be non-empty"))
		})

		It("returns error if namespace is empty", func() {
			_, err := NewHTTPClient(HTTPClientOpts{URL: "https://test-url"}, logger)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected config server namespace to be non-empty"))
		})

		It("returns error if CA certificate is not valid", func() {
			clientOpts, _, server := BuildServer()
			defer server.Close()

			clientOpts.TLSCA = []byte("invalid")

			_, err := NewHTTPClient(clientOpts, logger)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Building config server CA certificate"))
		})

		It("returns error if x509 key pair is not valid", func() {
			_, err := NewHTTPClient(HTTPClientOpts{URL: "https://test-url", Namespace: "test-ns"}, logger)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected non-empty config server authentication details"))
		})

		It("returns error if TLS cannot be verified", func() {
			clientOpts, certs, server := BuildServer()
			defer server.Close()

			clientOpts.TLSCA = []byte(certs.CA2.Certificate) // mismatch ca

			client, err := NewHTTPClient(clientOpts, logger)
			Expect(err).ToNot(HaveOccurred())

			_, err = client.Read("test-key")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("x509: certificate signed by unknown authority"))
		})

		It("returns error if UAA URL is specified without UAA client", func() {
			_, err := NewHTTPClient(HTTPClientOpts{
				URL:       "https://test-url",
				Namespace: "test-ns",

				UAAURL:    "https://uaa-url",
				UAAClient: "",
			}, logger)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Building config server UAA"))
		})
	})

	Describe("Read", func() {
		var (
			server *ghttp.Server
			client HTTPClient
		)

		BeforeEach(func() {
			var clientOpts HTTPClientOpts
			var err error
			clientOpts, _, server = BuildServer()
			client, err = NewHTTPClient(clientOpts, logger)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			server.Close()
		})

		It("succeeds making request and returns value", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v1/data", "name=test-ns/test-key"),
					ghttp.VerifyHeader(http.Header{"Accept": []string{"application/json"}}),
					ghttp.RespondWith(http.StatusOK, `{
						"data": [{
							"name": "test-key",
							"type": "test-type",
							"value": "test-value"
						}]
					}`),
				),
			)

			val, err := client.Read("test-key")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal("test-value"))
		})

		It("succeeds making request and returns complex hash value", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v1/data", "name=test-ns/test-key"),
					ghttp.VerifyHeader(http.Header{"Accept": []string{"application/json"}}),
					ghttp.RespondWith(http.StatusOK, `{
						"data": [{
							"name": "test-key",
							"type": "test-type",
							"value": {
								"test-hash-key": "test-hash-value",
								"test-array": [{
									"test-hash-key": "test-hash-value"
								}]
							}
						}]
					}`),
				),
			)

			val, err := client.Read("test-key")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(map[interface{}]interface{}{
				"test-hash-key": "test-hash-value",
				"test-array": []interface{}{
					map[interface{}]interface{}{
						"test-hash-key": "test-hash-value",
					},
				},
			}))
		})

		It("returns an error if response does not contain single value", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v1/data", "name=test-ns/test-key"),
					ghttp.VerifyHeader(http.Header{"Accept": []string{"application/json"}}),
					ghttp.RespondWith(http.StatusOK, `{"data": []}`),
				),
			)

			_, err := client.Read("test-key")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected to find at least one config server value for 'test-ns/test-key'"))
		})

		It("returns an error if response is not 200 ok", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v1/data", "name=test-ns/test-key"),
					ghttp.VerifyHeader(http.Header{"Accept": []string{"application/json"}}),
					ghttp.RespondWith(http.StatusNotFound, ``),
				),
			)

			_, err := client.Read("test-key")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Reading config server value 'test-ns/test-key'"))
		})
	})

	Describe("Read with UAA client", func() {
		var (
			server *ghttp.Server
			client HTTPClient
		)

		BeforeEach(func() {
			var clientOpts HTTPClientOpts
			var err error

			clientOpts, _, server = BuildServer()

			clientOpts.UAAURL = server.URL()
			clientOpts.UAAClient = "uaa-client"
			clientOpts.UAAClientSecret = "uaa-client-secret"

			client, err = NewHTTPClient(clientOpts, logger)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			server.Close()
		})

		It("succeeds doing a read with UAA configuration", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/oauth/token"),
					ghttp.VerifyBody([]byte("grant_type=client_credentials")),
					ghttp.VerifyBasicAuth("uaa-client", "uaa-client-secret"),
					ghttp.VerifyHeader(http.Header{"Content-Type": []string{"application/x-www-form-urlencoded"}}),
					ghttp.RespondWith(http.StatusOK, `{"access_token": "bearer uaa-token"}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v1/data", "name=test-ns/test-key"),
					ghttp.VerifyHeader(http.Header{"Authorization": []string{"bearer uaa-token"}}),
					ghttp.VerifyHeader(http.Header{"Accept": []string{"application/json"}}),
					ghttp.RespondWith(http.StatusOK, `{
						"data": [{
							"name": "test-key",
							"type": "test-type",
							"value": "test-value"
						}]
					}`),
				),
			)

			val, err := client.Read("test-key")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal("test-value"))
		})
	})

	Describe("Exists", func() {
		var (
			server *ghttp.Server
			client HTTPClient
		)

		BeforeEach(func() {
			var clientOpts HTTPClientOpts
			var err error
			clientOpts, _, server = BuildServer()
			client, err = NewHTTPClient(clientOpts, logger)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			server.Close()
		})

		It("succeeds making request and returns true if value exists", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v1/data", "name=test-ns/test-key"),
					ghttp.VerifyHeader(http.Header{"Accept": []string{"application/json"}}),
					ghttp.RespondWith(http.StatusOK, `{
						"data": [{
							"name": "test-key",
							"type": "test-type",
							"value": "test-value"
						}]
					}`),
				),
			)

			exists, err := client.Exists("test-key")
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeTrue())
		})

		It("succeeds making request and returns false if server returns 404", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v1/data", "name=test-ns/test-key"),
					ghttp.VerifyHeader(http.Header{"Accept": []string{"application/json"}}),
					ghttp.RespondWith(http.StatusNotFound, ``),
				),
			)

			exists, err := client.Exists("test-key")
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		It("succeeds making request and returns false if no value exists", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v1/data", "name=test-ns/test-key"),
					ghttp.VerifyHeader(http.Header{"Accept": []string{"application/json"}}),
					ghttp.RespondWith(http.StatusOK, `{
						"data": []
					}`),
				),
			)

			exists, err := client.Exists("test-key")
			Expect(err).ToNot(HaveOccurred())
			Expect(exists).To(BeFalse())
		})

		It("returns an error if response is not 200 or 404", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v1/data", "name=test-ns/test-key"),
					ghttp.VerifyHeader(http.Header{"Accept": []string{"application/json"}}),
					ghttp.RespondWith(http.StatusConflict, ``),
				),
			)

			_, err := client.Exists("test-key")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Reading config server value 'test-ns/test-key'"))
		})

		It("returns an error if response was not returned from the server", func() {
			terminateHttpConnection := func(w http.ResponseWriter, req *http.Request) {
				conn, _, _ := w.(http.Hijacker).Hijack()
				conn.Close()
			}

			server.AppendHandlers(
				// mutliple requests are added since client retries
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v1/data", "name=test-ns/test-key"),
					ghttp.VerifyHeader(http.Header{"Accept": []string{"application/json"}}),
					terminateHttpConnection,
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v1/data", "name=test-ns/test-key"),
					ghttp.VerifyHeader(http.Header{"Accept": []string{"application/json"}}),
					terminateHttpConnection,
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v1/data", "name=test-ns/test-key"),
					ghttp.VerifyHeader(http.Header{"Accept": []string{"application/json"}}),
					terminateHttpConnection,
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/api/v1/data", "name=test-ns/test-key"),
					ghttp.VerifyHeader(http.Header{"Accept": []string{"application/json"}}),
					terminateHttpConnection,
				),
			)

			exists, err := client.Exists("test-key")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Reading config server value 'test-ns/test-key'"))
			Expect(exists).To(BeFalse())
		})
	})

	Describe("Write", func() {
		var (
			server *ghttp.Server
			client HTTPClient
		)

		BeforeEach(func() {
			var clientOpts HTTPClientOpts
			var err error
			clientOpts, _, server = BuildServer()
			client, err = NewHTTPClient(clientOpts, logger)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			server.Close()
		})

		It("succeeds making request", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v1/data"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"application/json"},
						"Accept":       []string{"application/json"},
					}),
					ghttp.VerifyJSON(`{
						"name": "test-ns/test-key",
						"type": "value",
						"mode": "overwrite",
						"value": "test-value"
					}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := client.Write("test-key", "test-value")
			Expect(err).ToNot(HaveOccurred())
		})

		It("succeeds making request even for hashes with interface keys (json serialization interop)", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v1/data"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"application/json"},
						"Accept":       []string{"application/json"},
					}),
					ghttp.VerifyJSON(`{
						"name": "test-ns/test-key",
						"type": "value",
						"mode": "overwrite",
						"value": {
							"test-hash-key": "test-hash-value",
							"test-array": [{
								"test-hash-key": "test-hash-value"
							}]
						}
					}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := client.Write("test-key", map[interface{}]interface{}{
				"test-hash-key": "test-hash-value",
				"test-array": []interface{}{
					map[interface{}]interface{}{
						"test-hash-key": "test-hash-value",
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("succeeds making request even for []byte arrays (convert to string value since json does not have []byte type)", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v1/data"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"application/json"},
						"Accept":       []string{"application/json"},
					}),
					ghttp.VerifyJSON(`{
						"name": "test-ns/test-key",
						"type": "value",
						"mode": "overwrite",
						"value": "byte-array"
					}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			err := client.Write("test-key", []byte("byte-array"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error if serializing value fails", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v1/data"),
				),
			)

			err := client.Write("test-key", func() {})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Marshaling config server value"))
		})

		It("returns an error if response is not 200", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/api/v1/data"),
					ghttp.RespondWith(http.StatusConflict, ``),
				),
			)

			err := client.Write("test-key", "test-value")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Writing config server value 'test-ns/test-key'"))
		})
	})

	Describe("Delete", func() {
		var (
			server *ghttp.Server
			client HTTPClient
		)

		BeforeEach(func() {
			var clientOpts HTTPClientOpts
			var err error
			clientOpts, _, server = BuildServer()
			client, err = NewHTTPClient(clientOpts, logger)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			Expect(len(server.ReceivedRequests())).To(BeNumerically(">=", 1))
			server.Close()
		})

		It("succeeds making request", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/api/v1/data", "name=test-ns/test-key"),
					ghttp.VerifyHeader(http.Header{"Accept": []string{"application/json"}}),
					ghttp.RespondWith(http.StatusNoContent, ``),
				),
			)

			err := client.Delete("test-key")
			Expect(err).ToNot(HaveOccurred())
		})

		It("succeeds if response is 404", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/api/v1/data", "name=test-ns/test-key"),
					ghttp.VerifyHeader(http.Header{"Accept": []string{"application/json"}}),
					ghttp.RespondWith(http.StatusNotFound, ``),
				),
			)

			err := client.Delete("test-key")
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns an error if response is not 200 ok", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/api/v1/data", "name=test-ns/test-key"),
					ghttp.VerifyHeader(http.Header{"Accept": []string{"application/json"}}),
					ghttp.RespondWith(http.StatusInternalServerError, ``),
				),
			)

			err := client.Delete("test-key")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Deleting config server value 'test-ns/test-key'"))
		})
	})

	Describe("Generate", func() {
		var (
			server *ghttp.Server
			client HTTPClient
		)

		BeforeEach(func() {
			var clientOpts HTTPClientOpts
			var err error
			clientOpts, _, server = BuildServer()
			client, err = NewHTTPClient(clientOpts, logger)
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			server.Close()
		})

		It("succeeds making request", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v1/data"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"application/json"},
						"Accept":       []string{"application/json"},
					}),
					ghttp.VerifyJSON(`{
						"name": "test-ns/test-key",
						"type": "test-type",
						"parameters": "test-params"
					}`),
					ghttp.RespondWith(http.StatusOK, `{
						"value": "test-generated-value"
					}`),
				),
			)

			val, err := client.Generate("test-key", "test-type", "test-params")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal("test-generated-value"))
		})

		It("succeeds making request even for params that are hashes with interface keys (json serialization interop)", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v1/data"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"application/json"},
						"Accept":       []string{"application/json"},
					}),
					ghttp.VerifyJSON(`{
						"name": "test-ns/test-key",
						"type": "test-type",
						"parameters": {
							"test-hash-key": "test-hash-value",
							"test-array": [{
								"test-hash-key": "test-hash-value"
							}]
						}
					}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			_, err := client.Generate("test-key", "test-type", map[interface{}]interface{}{
				"test-hash-key": "test-hash-value",
				"test-array": []interface{}{
					map[interface{}]interface{}{
						"test-hash-key": "test-hash-value",
					},
				},
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("namespaces ca in parameters", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v1/data"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"application/json"},
						"Accept":       []string{"application/json"},
					}),
					ghttp.VerifyJSON(`{
						"name": "test-ns/test-key",
						"type": "test-type",
						"parameters": {
							"ca": "test-ns/test-ca"
						}
					}`),
					ghttp.RespondWith(http.StatusOK, `{}`),
				),
			)

			_, err := client.Generate("test-key", "test-type", map[interface{}]interface{}{
				"ca": "test-ca",
			})
			Expect(err).ToNot(HaveOccurred())
		})

		It("succeeds making request and returns complex hash", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v1/data"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"application/json"},
						"Accept":       []string{"application/json"},
					}),
					ghttp.VerifyJSON(`{
						"name": "test-ns/test-key",
						"type": "test-type",
						"parameters": "test-params"
					}`),
					ghttp.RespondWith(http.StatusOK, `{
						"value": {
							"test-hash-key": "test-hash-value",
							"test-array": [{
								"test-hash-key": "test-hash-value"
							}]
						}
					}`),
				),
			)

			val, err := client.Generate("test-key", "test-type", "test-params")
			Expect(err).ToNot(HaveOccurred())
			Expect(val).To(Equal(map[interface{}]interface{}{
				"test-hash-key": "test-hash-value",
				"test-array": []interface{}{
					map[interface{}]interface{}{
						"test-hash-key": "test-hash-value",
					},
				},
			}))
		})

		It("returns an error if serializing value fails", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v1/data"),
				),
			)

			_, err := client.Generate("test-key", "test-type", func() {})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Marshaling config server value"))
		})

		It("returns an error if response is not 200", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/api/v1/data"),
					ghttp.RespondWith(http.StatusConflict, ``),
				),
			)

			_, err := client.Generate("test-key", "test-type", "test-value")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Generating config server value 'test-ns/test-key'"))
		})
	})
})
