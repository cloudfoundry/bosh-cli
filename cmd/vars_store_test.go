package cmd_test

import (
	"errors"

	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
)

var _ = Describe("VarsStore", func() {
	var (
		fs    *fakesys.FakeFileSystem
		store *VarsStore
	)

	BeforeEach(func() {
		fs = fakesys.NewFakeFileSystem()
		store = &VarsStore{FS: fs}
	})

	Describe("Get", func() {
		Context("when delegated to scheme matching store", func() {
			BeforeEach(func() {
				err := store.UnmarshalFlag("fake-schema:///file")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns list results and errors from matched store", func() {
				store.RegisterSchemas(map[string]boshtpl.Variables{
					"fake-sch": &FakeVariables{
						GetResult: "fake-wrong-var",
						GetFound:  true,
						GetErr:    errors.New("fake-wrong-err"),
					},
					"fake-schema": &FakeVariables{
						GetResult: "fake-var",
						GetFound:  true,
						GetErr:    errors.New("fake-err"),
					},
				})

				val, found, err := store.Get(boshtpl.VariableDefinition{Name: "key"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
				Expect(found).To(BeTrue())
				Expect(val).To(Equal("fake-var"))
			})
		})

		Context("when delegated to FS store", func() {
			BeforeEach(func() {
				err := store.UnmarshalFlag("/file")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns value and found if store finds variable", func() {
				fs.WriteFileString("/file", "key: val")

				val, found, err := store.Get(boshtpl.VariableDefinition{Name: "key"})
				Expect(val).To(Equal("val"))
				Expect(found).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns error if reading file fails", func() {
				fs.WriteFileString("/file", "contents")
				fs.ReadFileError = errors.New("fake-err")

				_, _, err := store.Get(boshtpl.VariableDefinition{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})
	})

	Describe("List", func() {
		Context("when delegated to scheme matching store", func() {
			BeforeEach(func() {
				err := store.UnmarshalFlag("fake-schema:///file")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns list results and errors from matched store", func() {
				store.RegisterSchemas(map[string]boshtpl.Variables{
					"fake-sch": &FakeVariables{
						ListResults: []boshtpl.VariableDefinition{{Name: "fake-wrong-var-def"}},
						ListErr:     errors.New("fake-wrong-err"),
					},
					"fake-schema": &FakeVariables{
						ListResults: []boshtpl.VariableDefinition{{Name: "fake-var-def"}},
						ListErr:     errors.New("fake-err"),
					},
				})

				defs, err := store.List()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
				Expect(defs).To(Equal([]boshtpl.VariableDefinition{{Name: "fake-var-def"}}))
			})
		})

		Context("when delegated to FS store", func() {
			BeforeEach(func() {
				err := store.UnmarshalFlag("/file")
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns list of names without considering nested keys", func() {
				fs.WriteFileString("/file", "key1: val\nkey2: {key3: nested}")

				defs, err := store.List()
				Expect(defs).To(ConsistOf([]boshtpl.VariableDefinition{{Name: "key1"}, {Name: "key2"}}))
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns error if reading file fails", func() {
				fs.WriteFileString("/file", "contents")
				fs.ReadFileError = errors.New("fake-err")

				_, err := store.List()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})
	})

	Describe("IsSet", func() {
		It("returns true if store is configured with file path", func() {
			err := store.UnmarshalFlag("/file")
			Expect(err).ToNot(HaveOccurred())
			Expect(store.IsSet()).To(BeTrue())
		})

		It("returns false if store is not configured", func() {
			Expect(store.IsSet()).To(BeFalse())
		})
	})

	Describe("UnmarshalFlag", func() {
		It("returns error if file path is empty", func() {
			err := store.UnmarshalFlag("")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Expected file path to be non-empty"))
		})

		It("returns error if path cannot be expanded", func() {
			fs.ExpandPathErr = errors.New("fake-err")

			err := store.UnmarshalFlag("/file")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
