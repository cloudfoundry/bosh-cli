package director_test

import (
	"net/http"
	"time"

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

	Describe("OrphanNetworks", func() {
		It("returns orphaned networks", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/networks", "orphaned=true"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[
	{
		"name": "nw-1",
		"type": "manual",
		"created_at": "2015-01-09 06:23:25 +0000",
		"orphaned_at": "2016-01-09 06:23:25 +0000"
	},
	{
		"name": "nw-2",
		"type": "manual",
		"created_at": "2010-01-09 06:23:25 +0000",
		"orphaned_at": "2011-08-25 00:17:16 UTC"
	}
]`),
				),
			)

			networks, err := director.OrphanNetworks()
			Expect(err).ToNot(HaveOccurred())
			Expect(networks).To(HaveLen(2))

			Expect(networks[0].Name()).To(Equal("nw-1"))
			Expect(networks[0].Type()).To(Equal("manual"))
			Expect(networks[0].CreatedAt()).To(Equal(time.Date(2015, time.January, 9, 6, 23, 25, 0, time.UTC)))
			Expect(networks[0].OrphanedAt()).To(Equal(time.Date(2016, time.January, 9, 6, 23, 25, 0, time.UTC)))

			Expect(networks[1].Name()).To(Equal("nw-2"))
			Expect(networks[1].Type()).To(Equal("manual"))
			Expect(networks[1].CreatedAt()).To(Equal(time.Date(2010, time.January, 9, 6, 23, 25, 0, time.UTC)))
			Expect(networks[1].OrphanedAt()).To(Equal(time.Date(2011, time.August, 25, 0, 17, 16, 0, time.UTC)))
		})

		It("returns error if response is non-200", func() {
			AppendBadRequest(ghttp.VerifyRequest("GET", "/networks"), server)

			_, err := director.OrphanNetworks()
			Expect(err).To(HaveOccurred())
		})

		It("returns error if response cannot be unmarshalled", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/networks"),
					ghttp.RespondWith(http.StatusOK, ``),
				),
			)

			_, err := director.OrphanNetworks()
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("OrphanNetwork", func() {
	var (
		director Director
		network  OrphanNetwork
		server   *ghttp.Server
	)

	BeforeEach(func() {
		director, server = BuildServer()

		var err error

		network, err = director.FindOrphanNetwork("nw_name")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("Delete", func() {
		It("deletes orphaned network", func() {
			ConfigureTaskResult(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/networks/nw_name"),
					ghttp.VerifyBasicAuth("username", "password"),
				),
				"",
				server,
			)

			Expect(network.Delete()).ToNot(HaveOccurred())
		})

		It("succeeds even if error occurrs if network no longer exists", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/networks/nw_name"),
					ghttp.RespondWith(http.StatusBadRequest, ``),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/networks"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[]`),
				),
			)

			Expect(network.Delete()).ToNot(HaveOccurred())
		})

		It("returns delete error if listing networks fails", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/networks/nw_name"),
					ghttp.RespondWith(http.StatusBadRequest, ``),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/networks"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, ``),
				),
			)

			err := network.Delete()
			Expect(err).To(HaveOccurred())
		})

		It("returns error if response is non-200 and network still exists", func() {
			AppendBadRequest(ghttp.VerifyRequest("DELETE", "/networks/nw_name"), server)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/networks"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[
	{ "name": "nw_name", "orphaned_at": "2016-01-09 06:23:25 +0000" }
]`),
				),
			)

			err := network.Delete()
			Expect(err).To(HaveOccurred())
		})
	})
})
