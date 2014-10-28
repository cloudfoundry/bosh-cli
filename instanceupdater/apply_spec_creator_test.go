package instanceupdater_test

import (
	"errors"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	bmagentclient "github.com/cloudfoundry/bosh-micro-cli/agentclient"
	bmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell"

	. "github.com/cloudfoundry/bosh-micro-cli/instanceupdater"
)

var _ = Describe("ApplySpecCreator", func() {
	var (
		originalApplySpec bmstemcell.ApplySpec
		networksSpec      map[string]interface{}
		applySpecCreator  ApplySpecCreator
		fs                *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		originalApplySpec = bmstemcell.ApplySpec{
			Packages: map[string]bmstemcell.Blob{
				"fake-first-package-name": bmstemcell.Blob{
					Name: "fake-first-package-name",
				},
			},
			Job: bmstemcell.Job{
				Templates: []bmstemcell.Blob{
					{
						Name: "fake-template-name",
					},
				},
			},
		}

		networksSpec = map[string]interface{}{
			"fake-network-name": "fake-network-value",
		}

		fs = fakesys.NewFakeFileSystem()
		applySpecCreator = NewApplySpecCreator(fs)

		fs.RegisterOpenFile("/fake-archived-templates-path", &fakesys.FakeFile{
			Contents: []byte("fake-archive-contents"),
		})
		fs.RegisterOpenFile("/fake-templates-dir/file-1", &fakesys.FakeFile{
			Contents: []byte("fake-file-1-contents"),
		})
		fs.WriteFileString("/fake-templates-dir/file-1", "fake-file-1-contents")

		fs.RegisterOpenFile("/fake-templates-dir/config/file-2", &fakesys.FakeFile{
			Contents: []byte("fake-file-2-contents"),
		})
		fs.MkdirAll("/fake-templates-dir/config", os.ModePerm)
		fs.WriteFileString("/fake-templates-dir/config/file-2", "fake-file-2-contents")
	})

	Describe("Create", func() {
		It("creates an apply spec", func() {
			applySpec, err := applySpecCreator.Create(
				originalApplySpec,
				"fake-deployment-name",
				"fake-job-name",
				networksSpec,
				"fake-archived-templates-blob-id",
				"/fake-archived-templates-path",
				"/fake-templates-dir",
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(applySpec).To(Equal(bmagentclient.ApplySpec{
				Deployment: "fake-deployment-name",
				Index:      0,
				Packages: map[string]bmagentclient.Blob{
					"fake-first-package-name": bmagentclient.Blob{
						Name: "fake-first-package-name",
					},
				},
				ConfigurationHash: "bc0646cd41b98cd6c878db7a0573eca345f78200",
				Networks: map[string]interface{}{
					"fake-network-name": "fake-network-value",
				},
				Job: bmagentclient.Job{
					Name: "fake-job-name",
					Templates: []bmagentclient.Blob{
						{
							Name: "fake-template-name",
						},
					},
				},
				RenderedTemplatesArchive: bmagentclient.RenderedTemplatesArchiveSpec{
					BlobstoreID: "fake-archived-templates-blob-id",
					SHA1:        "4603db250d7b5b78dfe17869649784353177b549",
				},
			}))
		})

		Context("when creating the apply spec fails", func() {
			BeforeEach(func() {
				fs.OpenFileErr = errors.New("fake-open-file-error")
			})

			It("returns an error", func() {
				_, err := applySpecCreator.Create(
					originalApplySpec,
					"fake-deployment-name",
					"fake-job-name",
					networksSpec,
					"fake-archived-templates-blob-id",
					"/fake-archived-templates-path",
					"/fake-templates-dir",
				)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-open-file-error"))
			})
		})
	})
})
