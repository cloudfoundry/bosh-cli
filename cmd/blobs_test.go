package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	boshreldir "github.com/cloudfoundry/bosh-init/releasedir"
	fakereldir "github.com/cloudfoundry/bosh-init/releasedir/fakes"
	fakeui "github.com/cloudfoundry/bosh-init/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-init/ui/table"
)

var _ = Describe("BlobsCmd", func() {
	var (
		blobsDir *fakereldir.FakeBlobsDir
		ui       *fakeui.FakeUI
		command  BlobsCmd
	)

	BeforeEach(func() {
		blobsDir = &fakereldir.FakeBlobsDir{}
		ui = &fakeui.FakeUI{}
		command = NewBlobsCmd(blobsDir, ui)
	})

	Describe("Run", func() {
		act := func() error { return command.Run() }

		It("lists blobs", func() {
			blobs := []boshreldir.Blob{
				boshreldir.Blob{
					Path: "fake-path",
					Size: 100,

					BlobstoreID: "fake-blob-id",
					SHA1:        "fake-sha1",
				},
				boshreldir.Blob{
					Path: "dir/fake-path",
					Size: 1000,

					BlobstoreID: "",
					SHA1:        "fake-sha2",
				},
			}

			blobsDir.BlobsReturns(blobs, nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(ui.Table).To(Equal(boshtbl.Table{
				Content: "blobs",

				Header: []string{"Path", "Size", "Blobstore ID", "SHA1"},

				SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},

				Rows: [][]boshtbl.Value{
					{
						boshtbl.ValueString{"fake-path"},
						boshtbl.ValueBytes{100},
						boshtbl.ValueString{"fake-blob-id"},
						boshtbl.ValueString{"fake-sha1"},
					},
					{
						boshtbl.ValueString{"dir/fake-path"},
						boshtbl.ValueBytes{1000},
						boshtbl.ValueString{"(local)"},
						boshtbl.ValueString{"fake-sha2"},
					},
				},
			}))
		})

		It("returns error if blobs cannot be retrieved", func() {
			blobsDir.BlobsReturns(nil, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
