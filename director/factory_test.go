package director_test

import (
	"crypto/tls"
	"net/http"

	. "github.com/cloudfoundry/bosh-cli/director"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-utils/logger/loggerfakes"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	"github.com/onsi/gomega/ghttp"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

var _ = Describe("Factory", func() {
	Describe("New", func() {
		It("returns error if config is invalid", func() {
			_, err := NewFactory(boshlog.NewLogger(boshlog.LevelNone)).New(FactoryConfig{}, nil, nil)
			Expect(err).To(HaveOccurred())
		})

		It("returns error if TLS cannot be verified", func() {
			server := ghttp.NewTLSServer()
			defer server.Close()

			factoryConfig, err := NewConfigFromURL(server.URL())
			Expect(err).ToNot(HaveOccurred())

			logger := boshlog.NewLogger(boshlog.LevelNone)

			director, err := NewFactory(logger).New(factoryConfig, nil, nil)
			Expect(err).ToNot(HaveOccurred())

			_, err = director.Info()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("x509: certificate signed by unknown authority"))
		})

		Context("with valid TLS server", func() {
			var (
				server *ghttp.Server
			)

			BeforeEach(func() {
				server = ghttp.NewUnstartedServer()

				cert, err := tls.X509KeyPair(validCert, validKey)
				Expect(err).ToNot(HaveOccurred())

				server.HTTPTestServer.TLS = &tls.Config{
					Certificates: []tls.Certificate{cert},
				}

				server.HTTPTestServer.StartTLS()
			})

			AfterEach(func() {
				server.Close()
			})

			DirectorRedirect := func(config FactoryConfig) http.Header {
				h := http.Header{}
				// URL does not include port, creds
				h.Add("Location", "https://"+config.Host+"/info")
				h.Add("Referer", "referer")
				return h
			}

			TasksRedirect := func(config FactoryConfig) http.Header {
				h := http.Header{}
				// URL does not include port, creds
				h.Add("Location", "https://"+config.Host+"/tasks/123")
				h.Add("Referer", "referer")
				return h
			}

			VerifyHeaderDoesNotExist := func(key string) http.HandlerFunc {
				cKey := http.CanonicalHeaderKey(key)
				return func(w http.ResponseWriter, req *http.Request) {
					for k, _ := range req.Header {
						Expect(k).ToNot(Equal(cKey), "Header '%s' must not exist", cKey)
					}
				}
			}

			It("succeeds making requests and follow redirects with basic auth creds", func() {
				factoryConfig, err := NewConfigFromURL(server.URL())
				Expect(err).ToNot(HaveOccurred())

				factoryConfig.Client = "username"
				factoryConfig.ClientSecret = "password"
				factoryConfig.CACert = validCACert

				logger := boshlog.NewLogger(boshlog.LevelNone)

				director, err := NewFactory(logger).New(factoryConfig, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/info"),
						ghttp.VerifyBasicAuth("username", "password"),
						ghttp.RespondWith(http.StatusFound, nil, DirectorRedirect(factoryConfig)),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/info"),
						ghttp.VerifyBasicAuth("username", "password"),
						VerifyHeaderDoesNotExist("Referer"),
						ghttp.RespondWith(http.StatusOK, `{}`),
					),
				)

				_, err = director.Info()
				Expect(err).ToNot(HaveOccurred())
			})

			It("succeeds making initial post request and clears out headers when redirecting to a get resource", func() {
				fs := fakesys.NewFakeFileSystem()

				factoryConfig, err := NewConfigFromURL(server.URL())
				Expect(err).ToNot(HaveOccurred())

				factoryConfig.Client = "username"
				factoryConfig.ClientSecret = "password"
				factoryConfig.CACert = validCACert

				logger := boshlog.NewLogger(boshlog.LevelNone)

				taskReporter := NewNoopTaskReporter()
				fileReporter := NewNoopFileReporter()
				director, err := NewFactory(logger).New(factoryConfig, taskReporter, fileReporter)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/stemcells"),
						ghttp.VerifyBasicAuth("username", "password"),
						ghttp.RespondWith(http.StatusFound, nil, TasksRedirect(factoryConfig)),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/tasks/123"),
						ghttp.VerifyBasicAuth("username", "password"),
						VerifyHeaderDoesNotExist("Content-Type"),
						ghttp.RespondWith(http.StatusOK, `{"id":123, "state":"done"}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/tasks/123"),
						ghttp.VerifyBasicAuth("username", "password"),
						ghttp.RespondWith(http.StatusOK, `{"id":123, "state":"done"}`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/tasks/123/output", "type=event"),
						ghttp.VerifyBasicAuth("username", "password"),
						ghttp.RespondWith(http.StatusOK, ``),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/tasks/123/output", "type=result"),
						ghttp.VerifyBasicAuth("username", "password"),
						ghttp.RespondWith(http.StatusOK, ""),
					),
				)

				err = director.UploadStemcellFile(fakesys.NewFakeFile("/some-path", fs), false)
				Expect(err).ToNot(HaveOccurred())
			})

			It("does not redact url query params", func() {
				logger := &loggerfakes.FakeLogger{}
				factoryConfig, err := NewConfigFromURL(server.URL())
				Expect(err).ToNot(HaveOccurred())

				factoryConfig.Client = "username"
				factoryConfig.ClientSecret = "password"
				factoryConfig.CACert = validCACert

				director, err := NewFactory(logger).New(factoryConfig, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/cloud_configs", "limit=1"),
						ghttp.VerifyBasicAuth("username", "password"),
						ghttp.RespondWith(http.StatusOK, `[]`),
					),
				)

				director.LatestCloudConfig()

				debugMsgs := []interface{}{}
				for i := 0; i < logger.DebugCallCount(); i++ {
					_, _, args := logger.DebugArgsForCall(i)
					if len(args) >= 1 {
						debugMsgs = append(debugMsgs, args[0])
					}
				}

				Expect(debugMsgs).To(ContainElement(ContainSubstring("/cloud_configs?limit=1")))
			})

			It("succeeds making requests and follow redirects with token", func() {
				factoryConfig, err := NewConfigFromURL(server.URL())
				Expect(err).ToNot(HaveOccurred())

				var tokenRetries []bool

				factoryConfig.TokenFunc = func(retried bool) (string, error) {
					tokenRetries = append(tokenRetries, retried)
					return "auth", nil
				}
				factoryConfig.CACert = validCACert

				logger := boshlog.NewLogger(boshlog.LevelNone)

				director, err := NewFactory(logger).New(factoryConfig, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/info"),
						ghttp.VerifyHeader(http.Header{"Authorization": []string{"auth"}}),
						ghttp.RespondWith(http.StatusFound, nil, DirectorRedirect(factoryConfig)),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/info"),
						ghttp.VerifyHeader(http.Header{"Authorization": []string{"auth"}}),
						VerifyHeaderDoesNotExist("Referer"),
						ghttp.RespondWith(http.StatusOK, `{}`),
					),
				)

				_, err = director.Info()
				Expect(err).ToNot(HaveOccurred())

				// First token is fetched without retrying,
				// and on first redirect we forcefully retry
				// since redirects are not currently retried.
				Expect(tokenRetries).To(Equal([]bool{false, true}))
			})

			It("succeeds making requests and follow redirects without any auth", func() {
				factoryConfig, err := NewConfigFromURL(server.URL())
				Expect(err).ToNot(HaveOccurred())

				factoryConfig.CACert = validCACert

				logger := boshlog.NewLogger(boshlog.LevelNone)

				director, err := NewFactory(logger).New(factoryConfig, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/info"),
						VerifyHeaderDoesNotExist("Authorization"),
						ghttp.RespondWith(http.StatusFound, nil, DirectorRedirect(factoryConfig)),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/info"),
						VerifyHeaderDoesNotExist("Authorization"),
						VerifyHeaderDoesNotExist("Referer"),
						ghttp.RespondWith(http.StatusOK, `{}`),
					),
				)

				_, err = director.Info()
				Expect(err).ToNot(HaveOccurred())
			})

			It("retries request 3 times if a StatusGatewayTimeout returned", func() {
				factoryConfig, err := NewConfigFromURL(server.URL())
				Expect(err).ToNot(HaveOccurred())

				factoryConfig.CACert = validCACert

				logger := boshlog.NewLogger(boshlog.LevelNone)

				director, err := NewFactory(logger).New(factoryConfig, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/info"),
						ghttp.RespondWith(http.StatusGatewayTimeout, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/info"),
						ghttp.RespondWith(http.StatusGatewayTimeout, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/info"),
						ghttp.RespondWith(http.StatusOK, "{}"),
					),
				)

				_, err = director.Info()
				Expect(err).ToNot(HaveOccurred())
				Expect(len(server.ReceivedRequests())).To(Equal(3))
			})

			It("does not retry on non-successful http status codes", func() {
				factoryConfig, err := NewConfigFromURL(server.URL())
				Expect(err).ToNot(HaveOccurred())

				factoryConfig.CACert = validCACert

				logger := boshlog.NewLogger(boshlog.LevelNone)

				director, err := NewFactory(logger).New(factoryConfig, nil, nil)
				Expect(err).ToNot(HaveOccurred())

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/info"),
						ghttp.RespondWith(http.StatusGatewayTimeout, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/info"),
						ghttp.RespondWith(http.StatusInternalServerError, nil),
					),
				)

				_, err = director.Info()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Fetching info: Director responded with non-successful status code"))
				Expect(len(server.ReceivedRequests())).To(Equal(2))
			})

		})
	})
})
