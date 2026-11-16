package releasedir_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/releasedir"
)

var _ = Describe("FSConfig", func() {
	var (
		fs     *fakesys.FakeFileSystem
		config FSConfig
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		config = NewFSConfig("/dir/public.yml", "/dir/private.yml", fs)
	})

	Describe("Name", func() {
		It("returns name from public config", func() {
			err := fs.WriteFileString("/dir/public.yml", "name: name")
			Expect(err).ToNot(HaveOccurred())

			name, err := config.Name()
			Expect(err).ToNot(HaveOccurred())
			Expect(name).To(Equal("name"))
		})

		It("returns final_name from public config", func() {
			err := fs.WriteFileString("/dir/public.yml", "final_name: name")
			Expect(err).ToNot(HaveOccurred())

			name, err := config.Name()
			Expect(err).ToNot(HaveOccurred())
			Expect(name).To(Equal("name"))
		})

		It("returns error if name and final_name are empty", func() {
			err := fs.WriteFileString("/dir/public.yml", "")
			Expect(err).ToNot(HaveOccurred())

			_, err = config.Name()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected non-empty 'name' in config '/dir/public.yml'"))
		})

		It("returns error if both name and final_name are non-empty", func() {
			err := fs.WriteFileString("/dir/public.yml", "final_name: name\nname: name")
			Expect(err).ToNot(HaveOccurred())

			_, err = config.Name()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected 'name' or 'final_name' but not both in config '/dir/public.yml'"))
		})

		It("returns error if cannot read public config", func() {
			err := fs.WriteFileString("/dir/public.yml", "-")
			Expect(err).ToNot(HaveOccurred())
			fs.RegisterReadFileError("/dir/public.yml", errors.New("fake-err"))

			_, err = config.Name()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if cannot unmarshal public config", func() {
			err := fs.WriteFileString("/dir/public.yml", "-")
			Expect(err).ToNot(HaveOccurred())

			_, err = config.Name()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("line 1"))
		})

		It("returns error if cannot read private config", func() {
			err := fs.WriteFileString("/dir/private.yml", "-")
			Expect(err).ToNot(HaveOccurred())
			fs.RegisterReadFileError("/dir/private.yml", errors.New("fake-err"))

			_, err = config.Name()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if cannot unmarshal private config", func() {
			err := fs.WriteFileString("/dir/private.yml", "-")
			Expect(err).ToNot(HaveOccurred())

			_, err = config.Name()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("line 1"))
		})
	})

	Describe("Blobstore", func() {
		It("returns blobstore type name from public config", func() {
			err := fs.WriteFileString("/dir/public.yml", "blobstore: {provider: provider}")
			Expect(err).ToNot(HaveOccurred())

			provider, opts, err := config.Blobstore()
			Expect(err).ToNot(HaveOccurred())
			Expect(provider).To(Equal("provider"))
			Expect(opts).To(Equal(map[string]interface{}{}))
		})

		It("returns error if blobstore provider is empty", func() {
			err := fs.WriteFileString("/dir/public.yml", "")
			Expect(err).ToNot(HaveOccurred())

			_, _, err = config.Blobstore()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected non-empty 'blobstore.provider' in config '/dir/public.yml'"))
		})

		It("returns blobstore type and options name from public config", func() {
			err := fs.WriteFileString("/dir/public.yml", "blobstore: {provider: provider, options: {opt1: val1}}")
			Expect(err).ToNot(HaveOccurred())

			provider, opts, err := config.Blobstore()
			Expect(err).ToNot(HaveOccurred())
			Expect(provider).To(Equal("provider"))
			Expect(opts).To(Equal(map[string]interface{}{"opt1": "val1"}))
		})

		It("returns blobstore type and options name from public config, merged with options from private config", func() {
			err := fs.WriteFileString("/dir/public.yml",
				"blobstore: {provider: provider, options: {opt1: val1, opt2: pub-val}}")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString("/dir/private.yml",
				"blobstore: {options: {opt2: priv-val}}")
			Expect(err).ToNot(HaveOccurred())

			provider, opts, err := config.Blobstore()
			Expect(err).ToNot(HaveOccurred())
			Expect(provider).To(Equal("provider"))
			Expect(opts).To(Equal(map[string]interface{}{"opt1": "val1", "opt2": "priv-val"}))
		})

		It("returns error if cannot read public config", func() {
			err := fs.WriteFileString("/dir/public.yml", "-")
			Expect(err).ToNot(HaveOccurred())
			fs.RegisterReadFileError("/dir/public.yml", errors.New("fake-err"))

			_, _, err = config.Blobstore()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if cannot unmarshal public config", func() {
			err := fs.WriteFileString("/dir/public.yml", "-")
			Expect(err).ToNot(HaveOccurred())

			_, _, err = config.Blobstore()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("line 1"))
		})

		It("returns error if cannot read private config", func() {
			err := fs.WriteFileString("/dir/private.yml", "-")
			Expect(err).ToNot(HaveOccurred())
			fs.RegisterReadFileError("/dir/private.yml", errors.New("fake-err"))

			_, _, err = config.Blobstore()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if cannot unmarshal private config", func() {
			err := fs.WriteFileString("/dir/private.yml", "-")
			Expect(err).ToNot(HaveOccurred())

			_, _, err = config.Blobstore()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("line 1"))
		})
	})

	Describe("SaveName", func() {
		It("writes new config with name if config does not exist", func() {
			err := config.SaveName("new-name")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/public.yml")).To(Equal("name: new-name\n"))
		})

		It("adds name to public config keeping other entries", func() {
			err := fs.WriteFileString("/dir/public.yml", "name: name")
			Expect(err).ToNot(HaveOccurred())

			err = config.SaveName("new-name")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/public.yml")).To(Equal("name: new-name\n"))
		})

		It("overwrites existing name in public config keeping other entries", func() {
			err := fs.WriteFileString("/dir/public.yml", "name: name\nblobstore: {provider: s3}")
			Expect(err).ToNot(HaveOccurred())

			err = config.SaveName("new-name")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/public.yml")).To(Equal(
				"name: new-name\nblobstore:\n  provider: s3\n"))
		})

		It("migrates final_name to name", func() {
			err := fs.WriteFileString("/dir/public.yml", "final_name: name")
			Expect(err).ToNot(HaveOccurred())

			err = config.SaveName("new-name")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/public.yml")).To(Equal("name: new-name\n"))
		})

		It("returns error if cannot read public config", func() {
			err := fs.WriteFileString("/dir/public.yml", "-")
			Expect(err).ToNot(HaveOccurred())
			fs.RegisterReadFileError("/dir/public.yml", errors.New("fake-err"))

			err = config.SaveName("new-name")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if cannot unmarshal public config", func() {
			err := fs.WriteFileString("/dir/public.yml", "-")
			Expect(err).ToNot(HaveOccurred())

			err = config.SaveName("new-name")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("expected a mapping"))
		})

		It("keeps unknown top-level config entries", func() {
			err := fs.WriteFileString("/dir/public.yml", "name: name\nmin_cli_version: 2.0.0\nno_compression: true")
			Expect(err).ToNot(HaveOccurred())

			err = config.SaveName("new-name")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/public.yml")).To(Equal(
				"name: new-name\nmin_cli_version: 2.0.0\nno_compression: true\n"))
		})

		It("returns error if cannot write public config", func() {
			fs.WriteFileError = errors.New("fake-err")

			err := config.SaveName("new-name")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})

	Describe("SaveBlobstore", func() {
		It("writes provider and options to public config", func() {
			err := config.SaveBlobstore("s3", map[string]interface{}{"bucket_name": "bucket"}, nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/public.yml")).To(Equal(`blobstore:
  provider: s3
  options:
    bucket_name: bucket
`))
		})

		It("keeps existing public config entries", func() {
			err := fs.WriteFileString("/dir/public.yml", "name: dir\nmin_cli_version: 2.0.0")
			Expect(err).ToNot(HaveOccurred())

			err = config.SaveBlobstore("local", map[string]interface{}{"blobstore_path": ".blobs"}, nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/public.yml")).To(Equal(`name: dir
min_cli_version: 2.0.0
blobstore:
  provider: local
  options:
    blobstore_path: .blobs
`))
		})

		It("merges new options into existing blobstore options", func() {
			err := fs.WriteFileString("/dir/public.yml",
				"blobstore:\n  provider: s3\n  options:\n    bucket_name: old\n    region: us-east-1\n")
			Expect(err).ToNot(HaveOccurred())

			err = config.SaveBlobstore("s3", map[string]interface{}{"bucket_name": "new"}, nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/public.yml")).To(Equal(`blobstore:
  provider: s3
  options:
    bucket_name: new
    region: us-east-1
`))
		})

		It("replaces options when the provider changes", func() {
			err := fs.WriteFileString("/dir/public.yml",
				"blobstore:\n  provider: local\n  options:\n    blobstore_path: .blobs\n")
			Expect(err).ToNot(HaveOccurred())

			err = config.SaveBlobstore("s3", map[string]interface{}{"bucket_name": "bucket"}, nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/public.yml")).To(Equal(`blobstore:
  provider: s3
  options:
    bucket_name: bucket
`))
		})

		It("replaces private options when the provider changes", func() {
			err := fs.WriteFileString("/dir/public.yml",
				"blobstore:\n  provider: s3\n  options:\n    bucket_name: old\n")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString("/dir/private.yml",
				"blobstore:\n  options:\n    access_key_id: old\n")
			Expect(err).ToNot(HaveOccurred())

			err = config.SaveBlobstore("gcs",
				map[string]interface{}{"bucket_name": "bucket"},
				map[string]interface{}{"credentials_json": "new"})
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/private.yml")).To(Equal(`blobstore:
  options:
    credentials_json: new
`))
		})

		It("clears private options when the provider changes and no private options are given", func() {
			err := fs.WriteFileString("/dir/public.yml",
				"blobstore:\n  provider: s3\n  options:\n    bucket_name: old\n")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString("/dir/private.yml",
				"blobstore:\n  options:\n    access_key_id: old\n")
			Expect(err).ToNot(HaveOccurred())

			err = config.SaveBlobstore("gcs", map[string]interface{}{"bucket_name": "bucket"}, nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/private.yml")).ToNot(ContainSubstring("access_key_id"))
		})

		It("does not rewrite the public config when the private config cannot be read", func() {
			err := fs.WriteFileString("/dir/public.yml",
				"blobstore:\n  provider: s3\n  options:\n    bucket_name: old\n")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString("/dir/private.yml",
				"blobstore:\n  options:\n    access_key_id: old\n")
			Expect(err).ToNot(HaveOccurred())

			fs.RegisterReadFileError("/dir/private.yml", errors.New("fake-err"))

			err = config.SaveBlobstore("gcs",
				map[string]interface{}{"bucket_name": "bucket"},
				map[string]interface{}{"credentials_json": "new"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			fs.UnregisterReadFileError("/dir/private.yml")

			Expect(fs.ReadFileString("/dir/public.yml")).To(ContainSubstring("provider: s3"))
			Expect(fs.ReadFileString("/dir/private.yml")).To(ContainSubstring("access_key_id: old"))
		})

		It("does not rewrite the public config when the private config cannot be written", func() {
			err := fs.WriteFileString("/dir/private.yml",
				"blobstore:\n  options:\n    access_key_id: old\n")
			Expect(err).ToNot(HaveOccurred())

			fs.WriteFileErrors["/dir/private.yml"] = errors.New("fake-err")

			err = config.SaveBlobstore("gcs",
				map[string]interface{}{"bucket_name": "bucket"},
				map[string]interface{}{"credentials_json": "new"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(fs.FileExists("/dir/public.yml")).To(BeFalse())
		})

		It("cleans up stale private options when retrying after a private config write failure", func() {
			err := fs.WriteFileString("/dir/public.yml",
				"blobstore:\n  provider: s3\n  options:\n    bucket_name: old\n")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString("/dir/private.yml",
				"blobstore:\n  options:\n    access_key_id: old\n")
			Expect(err).ToNot(HaveOccurred())

			fs.WriteFileErrors["/dir/private.yml"] = errors.New("fake-err")

			err = config.SaveBlobstore("gcs", map[string]interface{}{"bucket_name": "bucket"}, nil)
			Expect(err).To(HaveOccurred())

			delete(fs.WriteFileErrors, "/dir/private.yml")

			err = config.SaveBlobstore("gcs", map[string]interface{}{"bucket_name": "bucket"}, nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/public.yml")).To(ContainSubstring("provider: gcs"))
			Expect(fs.ReadFileString("/dir/private.yml")).ToNot(ContainSubstring("access_key_id"))
		})

		It("restores the previous private config when the public config cannot be written", func() {
			err := fs.WriteFileString("/dir/public.yml",
				"blobstore:\n  provider: s3\n  options:\n    bucket_name: old\n")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString("/dir/private.yml",
				"blobstore:\n  options:\n    access_key_id: old\n")
			Expect(err).ToNot(HaveOccurred())

			fs.WriteFileErrors["/dir/public.yml"] = errors.New("fake-err")

			err = config.SaveBlobstore("gcs",
				map[string]interface{}{"bucket_name": "bucket"},
				map[string]interface{}{"credentials_json": "new"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(fs.ReadFileString("/dir/public.yml")).To(ContainSubstring("provider: s3"))
			Expect(fs.ReadFileString("/dir/private.yml")).To(ContainSubstring("access_key_id: old"))
			Expect(fs.ReadFileString("/dir/private.yml")).ToNot(ContainSubstring("credentials_json: new"))

			delete(fs.WriteFileErrors, "/dir/public.yml")

			err = config.SaveBlobstore("gcs",
				map[string]interface{}{"bucket_name": "bucket"},
				map[string]interface{}{"credentials_json": "new"})
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/public.yml")).To(ContainSubstring("provider: gcs"))
			Expect(fs.ReadFileString("/dir/private.yml")).To(ContainSubstring("credentials_json: new"))
			Expect(fs.ReadFileString("/dir/private.yml")).ToNot(ContainSubstring("access_key_id"))
		})

		It("removes a newly created private config when the public config cannot be written", func() {
			err := fs.WriteFileString("/dir/public.yml",
				"blobstore:\n  provider: s3\n")
			Expect(err).ToNot(HaveOccurred())

			fs.WriteFileErrors["/dir/public.yml"] = errors.New("fake-err")

			err = config.SaveBlobstore("gcs",
				map[string]interface{}{"bucket_name": "bucket"},
				map[string]interface{}{"credentials_json": "new"})
			Expect(err).To(HaveOccurred())

			Expect(fs.ReadFileString("/dir/public.yml")).To(ContainSubstring("provider: s3"))
			Expect(fs.FileExists("/dir/private.yml")).To(BeFalse())
		})

		It("defaults a local blobstore path when none is given", func() {
			err := config.SaveBlobstore("local", nil, nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/public.yml")).To(Equal(`blobstore:
  provider: local
  options:
    blobstore_path: .blobs
`))
		})

		It("keeps an existing local blobstore path", func() {
			err := fs.WriteFileString("/dir/public.yml",
				"blobstore:\n  provider: local\n  options:\n    blobstore_path: /custom\n")
			Expect(err).ToNot(HaveOccurred())

			err = config.SaveBlobstore("local", nil, nil)
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/public.yml")).To(ContainSubstring("blobstore_path: /custom"))
		})

		It("writes private options to private config", func() {
			err := config.SaveBlobstore("s3",
				map[string]interface{}{"bucket_name": "bucket"},
				map[string]interface{}{"access_key_id": "key"})
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/private.yml")).To(Equal(`blobstore:
  options:
    access_key_id: key
`))
		})

		It("returns error if cannot write public config", func() {
			fs.WriteFileError = errors.New("fake-err")

			err := config.SaveBlobstore("s3", nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})

	Describe("NoCompression", func() {
		It("returns false when no_compression is not present in final.yml", func() {
			err := fs.WriteFileString("/dir/public.yml", "name: test-release")
			Expect(err).ToNot(HaveOccurred())

			noCompression := config.NoCompression()
			Expect(noCompression).To(BeFalse())
		})

		It("returns false when no_compression is explicitly false in final.yml", func() {
			err := fs.WriteFileString("/dir/public.yml", "name: test-release\nno_compression: false")
			Expect(err).ToNot(HaveOccurred())

			noCompression := config.NoCompression()
			Expect(noCompression).To(BeFalse())
		})

		It("returns true when no_compression is true in final.yml", func() {
			err := fs.WriteFileString("/dir/public.yml", "name: test-release\nno_compression: true")
			Expect(err).ToNot(HaveOccurred())

			noCompression := config.NoCompression()
			Expect(noCompression).To(BeTrue())
		})
	})
})
