package ssh_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	boshssh "github.com/cloudfoundry/bosh-cli/ssh"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

var _ = Describe("SSHHostBuilder", func() {
	var (
		deployment  *fakedir.FakeDeployment
		hostBuilder boshssh.HostBuilder
	)

	BeforeEach(func() {
		deployment = &fakedir.FakeDeployment{}
		hostBuilder = boshssh.NewHostBuilder()
	})

	Describe("BuildHost", func() {

		var (
			act      func() (boshdir.Host, error)
			slug     boshdir.AllOrInstanceGroupOrInstanceSlug
			username string
		)

		BeforeEach(func() {
			act = func() (boshdir.Host, error) {
				return hostBuilder.BuildHost(slug, username, func() (boshdir.Deployment, error) {
					return deployment, nil
				})
			}
			username = "vcap"
		})

		Context("host is provided", func() {
			BeforeEach(func() {
				slug, _ = boshdir.NewAllOrInstanceGroupOrInstanceSlugFromString("1.2.3.4")
			})
			AfterEach(func() {
				Expect(deployment.VMInfosCallCount()).To(Equal(0))
			})

			It("Connects as specified user", func() {
				host, err := act()
				Expect(err).ToNot(HaveOccurred())
				Expect(host).To(Equal(boshdir.Host{
					Job:       "",
					IndexOrID: "",
					Username:  "vcap",
					Host:      "1.2.3.4",
				}))
			})
		})

		Context("instance has active VM with IP", func() {
			BeforeEach(func() {
				index := 1
				active := true
				deployment.VMInfosReturns([]boshdir.VMInfo{
					{
						JobName: "group",
						ID:      "id",
						Index:   &index,
						Active:  &active,
						IPs:     []string{"2.3.4.5"},
					},
				}, nil)
			})

			Context("instance ID is provided", func() {
				BeforeEach(func() {
					slug, _ = boshdir.NewAllOrInstanceGroupOrInstanceSlugFromString("group/id")
				})

				It("Connects to first IP", func() {
					host, err := act()
					Expect(err).ToNot(HaveOccurred())
					Expect(host).To(Equal(boshdir.Host{
						Job:       "group",
						IndexOrID: "id",
						Username:  "vcap",
						Host:      "2.3.4.5",
					}))
				})
			})

			Context("instance index is provided", func() {
				BeforeEach(func() {
					slug, _ = boshdir.NewAllOrInstanceGroupOrInstanceSlugFromString("group/1")
				})

				It("Connects to first IP", func() {
					host, err := act()
					Expect(err).ToNot(HaveOccurred())
					Expect(host).To(Equal(boshdir.Host{
						Job:       "group",
						IndexOrID: "1",
						Username:  "vcap",
						Host:      "2.3.4.5",
					}))
				})
			})

			Context("instance group is provided", func() {
				BeforeEach(func() {
					slug, _ = boshdir.NewAllOrInstanceGroupOrInstanceSlugFromString("group")
				})

				Context("group has single VM", func() {

					It("Connects to first IP", func() {
						host, err := act()
						Expect(err).ToNot(HaveOccurred())
						Expect(host).To(Equal(boshdir.Host{
							Job:       "group",
							IndexOrID: "",
							Username:  "vcap",
							Host:      "2.3.4.5",
						}))
					})

				})

				Context("group has multiple active VMs", func() {

					BeforeEach(func() {
						index := 1
						active := true
						deployment.VMInfosReturns([]boshdir.VMInfo{
							{
								JobName: "group",
								ID:      "id",
								Index:   &index,
								Active:  &active,
								IPs:     []string{"2.3.4.5"},
							},
							{
								JobName: "group",
								ID:      "id2",
								Index:   &index,
								Active:  &active,
								IPs:     []string{"2.3.4.6"},
							},
						}, nil)
					})

					It("Returns error", func() {
						_, err := act()
						Expect(err).To(Equal(errors.New("Instance group refers to more than 1 VM")))
						Expect(deployment.VMInfosCallCount()).To(Equal(1))
					})

				})
			})
		})

		Context("instance has active VM with no IP", func() {
			BeforeEach(func() {
				index := 1
				active := true
				deployment.VMInfosReturns([]boshdir.VMInfo{
					{
						JobName: "group",
						ID:      "id",
						Index:   &index,
						Active:  &active,
						IPs:     nil,
						VMID:    "cid",
					},
				}, nil)
				slug, _ = boshdir.NewAllOrInstanceGroupOrInstanceSlugFromString("group/id")
			})

			It("Returns error", func() {
				_, err := act()
				Expect(err).To(Equal(errors.New("VM cid has no IP address")))
				Expect(deployment.VMInfosCallCount()).To(Equal(1))
			})
		})

		Context("instance is not active", func() {
			BeforeEach(func() {
				index := 1
				active := false
				deployment.VMInfosReturns([]boshdir.VMInfo{
					{
						JobName: "group",
						ID:      "id",
						Index:   &index,
						Active:  &active,
						IPs:     []string{"1.2.3.4"},
						VMID:    "cid",
					},
				}, nil)
				slug, _ = boshdir.NewAllOrInstanceGroupOrInstanceSlugFromString("group/id")
			})

			It("Returns error", func() {
				_, err := act()
				Expect(err).To(Equal(errors.New("Instance group/id has no active VM")))
				Expect(deployment.VMInfosCallCount()).To(Equal(1))
			})
		})

		Context("VM lookup fails", func() {
			BeforeEach(func() {
				deployment.VMInfosReturns(nil, errors.New("oops"))
				slug, _ = boshdir.NewAllOrInstanceGroupOrInstanceSlugFromString("group/id")
			})

			It("Returns error", func() {
				_, err := act()
				Expect(err).To(Equal(bosherr.ComplexError{Cause: errors.New("oops"), Err: errors.New("Finding VM for group/id")}))
				Expect(deployment.VMInfosCallCount()).To(Equal(1))
			})
		})
	})
})
