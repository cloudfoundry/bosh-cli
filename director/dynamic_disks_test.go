package director_test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
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

	Describe("DynamicDisks", func() {
		It("returns dynamic disks", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/dynamic_disks"),
					ghttp.VerifyBasicAuth("username", "password"),
					ghttp.RespondWith(http.StatusOK, `[
	{
		"name": "disk1",
		"disk_cid": "cid1",
		"deployment": "dep1",
		"instance": "instance1",
		"availability_zone": "az1",
		"size": 1000,
		"disk_pool_name": "small",
		"cpi": "cpi1"
	},
	{
		"name": "disk2",
		"disk_cid": "cid2",
		"deployment": "dep2",
		"instance": "instance2",
		"availability_zone": "az2",
		"size": 2000,
		"disk_pool_name": "large",
		"cpi": "cpi2"
	}
]`),
				),
			)

			disks, err := director.DynamicDisks()
			Expect(err).ToNot(HaveOccurred())
			Expect(disks).To(HaveLen(2))

			Expect(disks[0].Name()).To(Equal("disk1"))
			Expect(disks[0].DiskCID()).To(Equal("cid1"))
			Expect(disks[0].DeploymentName()).To(Equal("dep1"))
			Expect(disks[0].InstanceName()).To(Equal("instance1"))
			Expect(disks[0].AvailabilityZone()).To(Equal("az1"))
			Expect(disks[0].Size()).To(Equal(uint64(1000)))
			Expect(disks[0].DiskPoolName()).To(Equal("small"))
			Expect(disks[0].CPI()).To(Equal("cpi1"))

			Expect(disks[1].Name()).To(Equal("disk2"))
			Expect(disks[1].DiskCID()).To(Equal("cid2"))
			Expect(disks[1].DeploymentName()).To(Equal("dep2"))
			Expect(disks[1].InstanceName()).To(Equal("instance2"))
			Expect(disks[1].AvailabilityZone()).To(Equal("az2"))
			Expect(disks[1].Size()).To(Equal(uint64(2000)))
			Expect(disks[1].DiskPoolName()).To(Equal("large"))
			Expect(disks[1].CPI()).To(Equal("cpi2"))
		})

		It("returns error if response is non-200", func() {
			AppendBadRequest(ghttp.VerifyRequest("GET", "/dynamic_disks"), server)

			_, err := director.DynamicDisks()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Listing dynamic disks: Director responded with non-successful status code"))
		})

		It("returns error if response cannot be unmarshalled", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/dynamic_disks"),
					ghttp.RespondWith(http.StatusOK, ``),
				),
			)

			_, err := director.DynamicDisks()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Listing dynamic disks: Unmarshaling Director response"))
		})
	})

	Describe("DeleteDynamicDisk", func() {
		It("deletes dynamic disk by name", func() {
			ConfigureTaskResult(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", "/dynamic_disks/my-disk"),
					ghttp.VerifyBasicAuth("username", "password"),
				),
				"",
				server,
			)

			Expect(director.DeleteDynamicDisk("my-disk")).ToNot(HaveOccurred())
		})

		It("does url encoding for disk name", func() {
			var verifyRawPath = func(path string) http.HandlerFunc {
				return func(w http.ResponseWriter, req *http.Request) {
					Expect(req.RequestURI).To(Equal(path))
				}
			}

			ConfigureTaskResult(
				ghttp.CombineHandlers(
					verifyRawPath("/dynamic_disks/my%3Bdisk"),
					ghttp.VerifyRequest("DELETE", "/dynamic_disks/my;disk"),
					ghttp.VerifyBasicAuth("username", "password"),
				),
				"",
				server,
			)

			Expect(director.DeleteDynamicDisk("my;disk")).ToNot(HaveOccurred())
		})

		It("returns error without contacting the director if disk name is empty", func() {
			err := director.DeleteDynamicDisk("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected non-empty dynamic disk name"))
			Expect(server.ReceivedRequests()).To(BeEmpty())
		})

		It("returns error if response is non-200", func() {
			AppendBadRequest(ghttp.VerifyRequest("DELETE", "/dynamic_disks/my-disk"), server)

			err := director.DeleteDynamicDisk("my-disk")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				"Deleting dynamic disk 'my-disk': Director responded with non-successful status code"))
		})
	})
})
