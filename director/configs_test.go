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

	Describe("LatestConfig", func() {
		It("returns the latest config", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/configs", "type=my-type&name=my-name&limit=1&latest=true"),
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
						ghttp.VerifyRequest("GET", "/configs", "type=missing-type&limit=1&latest=true&name=default"),
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
				AppendBadRequest(ghttp.VerifyRequest("GET", "/configs", "type=fake-type&limit=1&latest=true&name=default"), server)

				_, err := director.LatestConfig("fake-type", "default")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(
					"Finding config: Director responded with non-successful status code"))
			})
		})
	})

	Describe("LatestConfigByID", func() {
		It("returns config by ID", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/configs/123"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `{"id": "123", "type": "my-type", "name": "default", "content": "1"}`),
				),
			)

			cc, err := director.LatestConfigByID("123")
			Expect(err).ToNot(HaveOccurred())
			Expect(cc).To(Equal(Config{ID: "123", Name: "default", Type: "my-type", CreatedAt: "", Content: "1"}))
		})

		Context("when there is no config and director responds with status 404", func() {
			It("returns 'no config' error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/configs/bla"),
						ghttp.VerifyBasicAuth("username", "password"),
						ghttp.RespondWith(http.StatusNotFound, ""),
					),
				)

				_, err := director.LatestConfigByID("bla")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("No config"))
			})
		})

		Context("when server returns an error", func() {
			It("returns error", func() {
				AppendBadRequest(ghttp.VerifyRequest("GET", "/configs/123"), server)

				_, err := director.LatestConfigByID("123")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(
					"Finding config: Director responded with non-successful status code"))
			})
		})
	})

	Describe("ListConfigs", func() {
		Context("when limit is one", func() {
			Context("when no additional filters are given", func() {
				It("uses no query params and returns list of config items", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/configs", "limit=1&latest=true"),
							ghttp.VerifyBasicAuth("username", "password"),
							ghttp.RespondWith(http.StatusOK, `[{"name": "first", "type": "my-type", "created_at": "some-date", "team": "team1"}]`),
						),
					)
					cc, err := director.ListConfigs(1, ConfigsFilter{})
					Expect(err).ToNot(HaveOccurred())
					Expect(cc).To(Equal([]Config{{Type: "my-type", Name: "first", CreatedAt: "some-date", Team: "team1"}}))
				})
			})

			Context("when additional filters are given", func() {
				It("uses them as query parameters and returns list of config items", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/configs", "limit=1&latest=true&name=first&type=my-type"),
							ghttp.VerifyBasicAuth("username", "password"),
							ghttp.RespondWith(http.StatusOK, `[{"name": "first", "type": "my-type"}]`),
						),
					)

					cc, err := director.ListConfigs(1, ConfigsFilter{Type: "my-type", Name: "first"})
					Expect(err).ToNot(HaveOccurred())
					Expect(cc).To(Equal([]Config{{Type: "my-type", Name: "first"}}))
				})
			})
		})

		Context("when limit is greater one", func() {
			It("returns a list of config items", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/configs", "limit=2&latest=false"),
						ghttp.VerifyBasicAuth("username", "password"),
						ghttp.RespondWith(http.StatusOK, `[{"id": "some-id", "name": "first", "type": "my-type"}, {"id": "some-other-id", "name": "first", "type": "my-type"}]`),
					),
				)

				cc, err := director.ListConfigs(2, ConfigsFilter{})
				Expect(err).ToNot(HaveOccurred())
				Expect(cc).To(Equal([]Config{{ID: "some-id", Type: "my-type", Name: "first"}, {ID: "some-other-id", Type: "my-type", Name: "first"}}))
			})
		})

		Context("when server returns an error", func() {
			It("returns error", func() {
				AppendBadRequest(ghttp.VerifyRequest("GET", "/configs"), server)

				_, err := director.ListConfigs(1, ConfigsFilter{})
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
					ghttp.VerifyRequest("POST", "/configs"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.VerifyBody([]byte(`{"type":"my-type","name":"my-name","content":"---"}`)),
					ghttp.VerifyHeader(http.Header{"Content-Type": []string{"application/json"}}),
					ghttp.RespondWith(http.StatusOK, []byte("{}")),
				),
			)

			_, err := director.UpdateConfig("my-type", "my-name", []byte("---"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("keeps yaml content intact", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/configs"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.VerifyBody([]byte(`{"type":"my-type","name":"my-name","content":"abc\ndef\n"}`)),
					ghttp.VerifyHeader(http.Header{"Content-Type": []string{"application/json"}}),
					ghttp.RespondWith(http.StatusOK, []byte("{}")),
				),
			)

			_, err := director.UpdateConfig("my-type", "my-name", []byte("abc\ndef\n"))
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns config", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/configs"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, []byte(`{"id":"123","type":"my-type","name":"my-name","created_at":"","content":"a"}`)),
				),
			)

			config, err := director.UpdateConfig("my-type", "my-name", []byte("abc\ndef\n"))
			Expect(err).ToNot(HaveOccurred())
			Expect(config).To(Equal(Config{ID: "123", Name: "my-name", Type: "my-type", CreatedAt: "", Content: "a"}))
		})

		Context("when server returns an error", func() {
			It("returns error", func() {
				AppendBadRequest(ghttp.VerifyRequest("POST", "/configs"), server)

				_, err := director.UpdateConfig("fake-type", "fake-name", nil)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(
					"Updating config: Director responded with non-successful status code '400'"))
			})

			It("returns error for invalid JSON", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/configs"),
						ghttp.VerifyBasicAuth("username", "password"),
						ghttp.RespondWith(http.StatusOK, []byte(`abc\ndef\n`)),
					),
				)

				config, err := director.UpdateConfig("my-type", "my-name", nil)
				Expect(err).To(HaveOccurred())
				Expect(config).To(Equal(Config{}))
				Expect(err.Error()).To(ContainSubstring("Unmarshaling Director response"))
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

	Describe("DiffConfig", func() {
		expectedDiffResponse := ConfigDiff{
			Diff: [][]interface{}{
				{"release:", nil},
				{"  version: 0.0.1", "removed"},
				{"  version: 0.0.2", "added"},
			},
		}

		It("diffs the config with the given name", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/configs/diff"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"application/json"},
					}),
					ghttp.VerifyBody([]byte(`{"type":"myType","name":"myName","content":"myConfig"}`)),
					ghttp.RespondWith(http.StatusOK, `{"diff":[["release:",null],["  version: 0.0.1","removed"],["  version: 0.0.2","added"]]}`),
				),
			)

			diff, err := director.DiffConfig("myType", "myName", []byte("myConfig"))
			Expect(err).ToNot(HaveOccurred())
			Expect(diff).To(Equal(expectedDiffResponse))
		})

		It("returns error if info response in non-200", func() {
			AppendBadRequest(ghttp.VerifyRequest("POST", "/configs/diff"), server)

			_, err := director.DiffConfig("myType", "myName", nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Fetching diff result: Director responded with non-successful status code"))
		})

	})
})
