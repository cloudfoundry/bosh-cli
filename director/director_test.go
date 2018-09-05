package director_test

import (
	"bytes"
	"net/http"
	"time"

	"github.com/cloudfoundry/bosh-cli/director"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Director", func() {
	var (
		dir    director.Director
		server *ghttp.Server
	)

	BeforeEach(func() {
		dir, server = BuildServer()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("OrphanedVMs", func() {
		It("returns parsed orphaned vms", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/orphaned_vms"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, []map[string]interface{}{
						{
							"az":              "az1",
							"cid":             "cid-1",
							"deployment_name": "d-1",
							"instance_name":   "i-1",
							"ip_addresses":    []string{"1.1.1.1", "2.2.2.2"},
							"orphaned_at":     "2020-04-03 08:08:08 UTC",
						},
						{
							"az":              "az2",
							"cid":             "cid-2",
							"deployment_name": "d-2",
							"instance_name":   "i-2",
							"ip_addresses":    []string{"3.3.3.3"},
							"orphaned_at":     "2021-06-04 08:08:08 UTC",
						},
					}),
				),
			)

			orphanedVMs, err := dir.OrphanedVMs()
			Expect(err).ToNot(HaveOccurred())
			Expect(orphanedVMs).To(ConsistOf(
				director.OrphanedVM{
					AZName:         "az1",
					CID:            "cid-1",
					DeploymentName: "d-1",
					InstanceName:   "i-1",
					IPAddresses:    []string{"1.1.1.1", "2.2.2.2"},
					OrphanedAt:     time.Date(2020, 04, 03, 8, 8, 8, 0, time.UTC),
				},
				director.OrphanedVM{
					AZName:         "az2",
					CID:            "cid-2",
					DeploymentName: "d-2",
					InstanceName:   "i-2",
					IPAddresses:    []string{"3.3.3.3"},
					OrphanedAt:     time.Date(2021, 06, 04, 8, 8, 8, 0, time.UTC),
				},
			))
		})

		It("returns error if response is non-200", func() {
			AppendBadRequest(ghttp.VerifyRequest("GET", "/orphaned_vms"), server)

			_, err := dir.OrphanedVMs()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Director responded with non-successful status code '400'"))
		})
	})

	Describe("EnableResurrection", func() {
		It("enables resurrection for all instances and returns without an error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/resurrection"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"application/json"},
					}),
					ghttp.VerifyBody([]byte(`{"resurrection_paused":false}`)),
				),
			)

			err := dir.EnableResurrection(true)
			Expect(err).ToNot(HaveOccurred())
		})

		It("disables resurrection for all instances and returns without an error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PUT", "/resurrection"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"application/json"},
					}),
					ghttp.VerifyBody([]byte(`{"resurrection_paused":true}`)),
				),
			)

			err := dir.EnableResurrection(false)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if response is non-200", func() {
			AppendBadRequest(ghttp.VerifyRequest("PUT", "/resurrection"), server)

			err := dir.EnableResurrection(true)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Changing VM resurrection state"))
		})
	})

	Describe("CleanUp", func() {
		It("cleans up all resources and returns without an error", func() {
			ConfigureTaskResult(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/cleanup"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"application/json"},
					}),
					ghttp.VerifyBody([]byte(`{"config":{"remove_all":true}}`)),
				),
				"",
				server,
			)

			err := dir.CleanUp(true)
			Expect(err).ToNot(HaveOccurred())
		})

		It("cleans up some resources and returns without an error", func() {
			ConfigureTaskResult(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/cleanup"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.VerifyHeader(http.Header{
						"Content-Type": []string{"application/json"},
					}),
					ghttp.VerifyBody([]byte(`{"config":{"remove_all":false}}`)),
				),
				"",
				server,
			)

			err := dir.CleanUp(false)
			Expect(err).ToNot(HaveOccurred())
		})

		It("returns error if response is non-200", func() {
			AppendBadRequest(ghttp.VerifyRequest("POST", "/cleanup"), server)

			err := dir.CleanUp(true)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Cleaning up resources"))
		})
	})

	Describe("DownloadResourceUnchecked", func() {
		var (
			buf *bytes.Buffer
		)

		BeforeEach(func() {
			buf = bytes.NewBufferString("")
		})

		It("writes to the writer downloaded contents and returns without an error", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/resources/blob-id"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, "result"),
				),
			)

			err := dir.DownloadResourceUnchecked("blob-id", buf)
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("result"))
		})

		It("returns error if response is non-200", func() {
			AppendBadRequest(ghttp.VerifyRequest("GET", "/resources/blob-id"), server)

			err := dir.DownloadResourceUnchecked("blob-id", buf)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Downloading resource 'blob-id'"))
		})
	})

	Describe("With Context", func() {
		It("Adds the context id to requests", func() {
			buf := bytes.NewBufferString("")
			contextId := "example-context-id"
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/resources/blob-id"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.VerifyHeaderKV("X-Bosh-Context-Id", contextId),
					ghttp.RespondWith(http.StatusOK, contextId),
				),
			)

			dir = dir.WithContext(contextId)
			err := dir.DownloadResourceUnchecked("blob-id", buf)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
