package director_test

import (
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

	Describe("Cleanup", func() {
		Context("when the dryrun flag is set to true", func() {
			It("shows what it would clean-up", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/cleanup/dryrun", "remove_all=false&keep_orphaned_disks=false"),
						ghttp.VerifyBasicAuth("username", "password"),
						ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]interface{}{
							"releases": []map[string]interface{}{
								{
									"name":     "releases-1",
									"versions": []string{"1"},
								},
								{
									"name":     "releases-2",
									"versions": []string{"1", "2"},
								},
							},
							"stemcells": []map[string]interface{}{
								{
									"name":    "bosh-warden-boshlite-ubuntu-xenial-go_agent",
									"version": "621.21",
								},
							},
							"compiled_packages": []map[string]string{
								{
									"package_name":     "packagename",
									"stemcell_version": "621.6",
									"stemcell_os":      "ubuntu-xenial",
								},
							},
							"orphaned_disks": []map[string]interface{}{
								{
									"deployment_name": "zookeeper",
									"size":            10240,
									"disk_cid":        "disk-cid",
									"az":              "az1",
									"instance_name":   "i-1",
								},
							},
							"orphaned_vms": []map[string]interface{}{
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
							},
							"exported_releases": []string{"release_blob_id"},
							"dns_blobs":         []string{"dns-blob-id1", "dns-blob-id2"},
						}),
					),
				)
				resp, err := dir.CleanUp(false, true, false)
				Expect(err).ToNot(HaveOccurred())

				Expect(resp.Releases).To(ConsistOf(
					director.CleanableRelease{
						Name:     "releases-1",
						Versions: []string{"1"},
					},
					director.CleanableRelease{
						Name:     "releases-2",
						Versions: []string{"1", "2"},
					},
				))

				Expect(resp.OrphanedVMs).To(ConsistOf(
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
		})
	})
})
