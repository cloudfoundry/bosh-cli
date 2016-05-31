package releasedir_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/releasedir"
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

	Describe("FinalName", func() {
		It("returns final name from public config", func() {
			fs.WriteFileString("/dir/public.yml", "final_name: name")

			name, err := config.FinalName()
			Expect(err).ToNot(HaveOccurred())
			Expect(name).To(Equal("name"))
		})

		It("returns error if name is empty", func() {
			fs.WriteFileString("/dir/public.yml", "")

			_, err := config.FinalName()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected non-empty 'final_name' in config '/dir/public.yml'"))
		})

		It("returns error if cannot read public config", func() {
			fs.WriteFileString("/dir/public.yml", "-")
			fs.RegisterReadFileError("/dir/public.yml", errors.New("fake-err"))

			_, err := config.FinalName()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if cannot unmarshal public config", func() {
			fs.WriteFileString("/dir/public.yml", "-")

			_, err := config.FinalName()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("line 1"))
		})

		It("returns error if cannot read private config", func() {
			fs.WriteFileString("/dir/private.yml", "-")
			fs.RegisterReadFileError("/dir/private.yml", errors.New("fake-err"))

			_, err := config.FinalName()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if cannot unmarshal private config", func() {
			fs.WriteFileString("/dir/private.yml", "-")

			_, err := config.FinalName()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("line 1"))
		})
	})

	Describe("Blobstore", func() {
		It("returns blobstore type name from public config", func() {
			fs.WriteFileString("/dir/public.yml", "blobstore: {provider: provider}")

			provider, opts, err := config.Blobstore()
			Expect(err).ToNot(HaveOccurred())
			Expect(provider).To(Equal("provider"))
			Expect(opts).To(Equal(map[string]interface{}{}))
		})

		It("returns error if blobstore provider is empty", func() {
			fs.WriteFileString("/dir/public.yml", "")

			_, _, err := config.Blobstore()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected non-empty 'blobstore.provider' in config '/dir/public.yml'"))
		})

		It("returns blobstore type and options name from public config", func() {
			fs.WriteFileString("/dir/public.yml", "blobstore: {provider: provider, options: {opt1: val1}}")

			provider, opts, err := config.Blobstore()
			Expect(err).ToNot(HaveOccurred())
			Expect(provider).To(Equal("provider"))
			Expect(opts).To(Equal(map[string]interface{}{"opt1": "val1"}))
		})

		It("returns blobstore type and options name from public config, merged with options from private config", func() {
			fs.WriteFileString("/dir/public.yml",
				"blobstore: {provider: provider, options: {opt1: val1, opt2: pub-val}}")

			fs.WriteFileString("/dir/private.yml",
				"blobstore: {options: {opt2: priv-val}}")

			provider, opts, err := config.Blobstore()
			Expect(err).ToNot(HaveOccurred())
			Expect(provider).To(Equal("provider"))
			Expect(opts).To(Equal(map[string]interface{}{"opt1": "val1", "opt2": "priv-val"}))
		})

		It("returns error if cannot read public config", func() {
			fs.WriteFileString("/dir/public.yml", "-")
			fs.RegisterReadFileError("/dir/public.yml", errors.New("fake-err"))

			_, _, err := config.Blobstore()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if cannot unmarshal public config", func() {
			fs.WriteFileString("/dir/public.yml", "-")

			_, _, err := config.Blobstore()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("line 1"))
		})

		It("returns error if cannot read private config", func() {
			fs.WriteFileString("/dir/private.yml", "-")
			fs.RegisterReadFileError("/dir/private.yml", errors.New("fake-err"))

			_, _, err := config.Blobstore()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if cannot unmarshal private config", func() {
			fs.WriteFileString("/dir/private.yml", "-")

			_, _, err := config.Blobstore()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("line 1"))
		})
	})

	Describe("SaveFinalName", func() {
		It("writes new config with name if config does not exist", func() {
			err := config.SaveFinalName("new-name")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/public.yml")).To(Equal("final_name: new-name\n"))
		})

		It("adds name to public config keeping other entries", func() {
			fs.WriteFileString("/dir/public.yml", "final_name: name")

			err := config.SaveFinalName("new-name")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/public.yml")).To(Equal("final_name: new-name\n"))
		})

		It("overwrites existing name in public config keeping other entries", func() {
			fs.WriteFileString("/dir/public.yml", "final_name: name\nblobstore: {provider: s3}")

			err := config.SaveFinalName("new-name")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/public.yml")).To(Equal(
				"final_name: new-name\nblobstore:\n  provider: s3\n"))
		})

		It("returns error if cannot read public config", func() {
			fs.WriteFileString("/dir/public.yml", "-")
			fs.RegisterReadFileError("/dir/public.yml", errors.New("fake-err"))

			err := config.SaveFinalName("new-name")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if cannot unmarshal public config", func() {
			fs.WriteFileString("/dir/public.yml", "-")

			err := config.SaveFinalName("new-name")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("line 1"))
		})

		It("returns error if cannot read private config", func() {
			fs.WriteFileString("/dir/private.yml", "-")
			fs.RegisterReadFileError("/dir/private.yml", errors.New("fake-err"))

			err := config.SaveFinalName("new-name")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if cannot unmarshal private config", func() {
			fs.WriteFileString("/dir/private.yml", "-")

			err := config.SaveFinalName("new-name")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("line 1"))
		})

		It("returns error if cannot write public config", func() {
			fs.WriteFileError = errors.New("fake-err")

			err := config.SaveFinalName("new-name")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
