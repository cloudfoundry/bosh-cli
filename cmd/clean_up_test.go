package cmd_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	. "github.com/cloudfoundry/bosh-cli/cmd/opts"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	"github.com/cppforlife/go-semi-semantic/version"
)

var _ = Describe("CleanUpCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  CleanUpCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = NewCleanUpCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			opts CleanUpOpts
		)

		BeforeEach(func() {
			opts = CleanUpOpts{}
		})

		act := func() error { return command.Run(opts) }

		It("cleans up director resources", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.CleanUpCallCount()).To(Equal(1))
			Expect(director.CleanUpArgsForCall(0)).To(BeFalse())
		})

		It("cleans up *all* director resources", func() {
			opts.All = true

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.CleanUpCallCount()).To(Equal(1))
			Expect(director.CleanUpArgsForCall(0)).To(BeTrue())
		})

		It("does not clean up if confirmation is rejected", func() {
			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(director.CleanUpCallCount()).To(Equal(0))
		})

		It("returns error if cleaning up fails", func() {
			director.CleanUpReturns(boshdir.CleanUp{}, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})

	Describe("Print", func() {
		BeforeEach(func() {
			disks := []boshdir.OrphanDisk{
				&fakedir.FakeOrphanDisk{
					CIDStub:  func() string { return "cid" },
					SizeStub: func() uint64 { return 100 },

					DeploymentStub: func() boshdir.Deployment {
						return &fakedir.FakeDeployment{
							NameStub: func() string { return "deployment" },
						}
					},
					InstanceNameStub: func() string { return "instance" },
					AZNameStub:       func() string { return "az" },

					OrphanedAtStub: func() time.Time {
						return time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
					},
				},
			}

			director.OrphanDisksReturns(disks, nil)
		})

		It("shows a big table of all artifacts that could be deleted", func() {
			releases := []boshdir.CleanableRelease{
				{Name: "release1", Versions: []string{"1.5.4"}},
				{Name: "release2", Versions: []string{"1.1.4"}},
			}

			stemcells := []boshdir.Stemcell{
				&fakedir.FakeStemcell{
					NameStub: func() string {
						return "bosh-warden"
					},
					VersionStub: func() version.Version {
						return version.MustNewVersionFromString("621.21")
					},
				},
			}

			compiledPackages := []boshdir.CleanableCompiledPackage{
				{Name: "bpm", StemcellOs: "ubuntu-xenial", StemcellVersion: "621.5"},
				{Name: "bpm", StemcellOs: "ubuntu-xenial", StemcellVersion: "456.20"},
			}

			exportedReleases := []string{"exported-release-blob-id"}
			dnsBlobs := []string{"dns-blob-id", "dns-blob-id1"}

			orphanedDisks := []boshdir.OrphanDiskResp{
				{CID: "disk-cid1", InstanceName: "some-instance/123", DeploymentName: "some-deployment", Size: 10240},
				{CID: "disk-cid2", InstanceName: "some-instance/123", DeploymentName: "some-deployment", Size: 20480},
				{CID: "disk-cid3", InstanceName: "some-instance/123", DeploymentName: "some-deployment", Size: 1337},
				{CID: "disk-cid4", InstanceName: "some-instance/123", DeploymentName: "some-deployment", Size: 13370},
			}

			orphanedVms := []boshdir.OrphanedVM{
				{CID: "vm-cid", InstanceName: "some-instance/123", DeploymentName: "some-deployment"},
			}

			stuff := boshdir.CleanUp{
				Releases:         releases,
				Stemcells:        stemcells,
				CompiledPackages: compiledPackages,
				ExportedReleases: exportedReleases,
				DNSBlobs:         dnsBlobs,
				OrphanedDisks:    orphanedDisks,
				OrphanedVMs:      orphanedVms,
			}

			command.PrintCleanUpTable(stuff)

			Expect(len(ui.Tables)).To(Equal(7))

			Expect(ui.Tables[0].Title).To(Equal("Unused Releases"))
			Expect(len(ui.Tables[0].Header)).To(Equal(2))
			Expect(len(ui.Tables[0].Rows)).To(Equal(2))

			Expect(ui.Tables[1].Title).To(Equal("Unused Stemcells"))
			Expect(len(ui.Tables[1].Header)).To(Equal(2))
			Expect(len(ui.Tables[1].Rows)).To(Equal(1))

			Expect(ui.Tables[2].Title).To(Equal("Unused Compiled Packages"))
			Expect(len(ui.Tables[2].Header)).To(Equal(3))
			Expect(len(ui.Tables[2].Rows)).To(Equal(2))

			Expect(ui.Tables[3].Title).To(Equal("Exported Releases"))
			Expect(len(ui.Tables[3].Header)).To(Equal(1))
			Expect(len(ui.Tables[3].Rows)).To(Equal(1))

			Expect(ui.Tables[4].Title).To(Equal("Stale DNS Record Blobs"))
			Expect(len(ui.Tables[4].Header)).To(Equal(1))
			Expect(len(ui.Tables[4].Rows)).To(Equal(2))

			Expect(ui.Tables[5].Title).To(Equal("Orphaned Disks"))
			Expect(len(ui.Tables[5].Header)).To(Equal(4))
			Expect(len(ui.Tables[5].Rows)).To(Equal(4))

			Expect(ui.Tables[6].Title).To(Equal("Orphaned VMs"))
			Expect(len(ui.Tables[6].Header)).To(Equal(3))
			Expect(len(ui.Tables[6].Rows)).To(Equal(1))
		})
	})
})
