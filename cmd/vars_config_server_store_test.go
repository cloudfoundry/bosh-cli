package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	bconfigserver "github.com/cloudfoundry/bosh-cli/configserver"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
)

var _ = Describe("VarsConfigServerStore", func() {
	var (
		client *bconfigserver.MemoryClient
		store  *VarsConfigServerStore
	)

	BeforeEach(func() {
		client = bconfigserver.NewMemoryClient()
		store = NewConfigServerVarsStore(client)
	})

	Describe("Get", func() {
		It("returns value and found if store finds variable", func() {
			client.Write("key", "val")

			val, found, err := store.Get(boshtpl.VariableDefinition{Name: "key"})
			Expect(val).To(Equal("val"))
			Expect(found).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when store does not find variable", func() {
			It("returns nil and not found if variable type is not available", func() {
				client.Write("key", "val")

				val, found, err := store.Get(boshtpl.VariableDefinition{Name: "key2"})
				Expect(val).To(BeNil())
				Expect(found).To(BeFalse())
				Expect(err).ToNot(HaveOccurred())
			})

			It("tries to generate value and save it if variable type is available", func() {
				client := &FakeConfigServerClient{}
				store = NewConfigServerVarsStore(client)

				client.GenerateResult = "password"

				val, found, err := store.Get(boshtpl.VariableDefinition{Name: "key2", Type: "password", Options: "options"})
				Expect(val.(string)).To(Equal("password"))
				Expect(found).To(BeTrue())
				Expect(err).ToNot(HaveOccurred())

				Expect(client.GenerateName).To(Equal("key2"))
				Expect(client.GenerateType).To(Equal("password"))
				Expect(client.GenerateParams).To(Equal("options"))
			})

			It("returns error if generating value fails", func() {
				client := &FakeConfigServerClient{}
				store = NewConfigServerVarsStore(client)

				client.GenerateErr = errors.New("fake-err")

				val, found, err := store.Get(boshtpl.VariableDefinition{Name: "key2", Type: "password"})
				Expect(val).To(BeNil())
				Expect(found).To(BeFalse())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})
	})

	Describe("List", func() {
		It("returns not implemented error", func() {
			_, err := store.List()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Listing of variables in config server is not supported"))
		})
	})
})

type FakeConfigServerClient struct {
	bconfigserver.ErrClient
	GenerateName   string
	GenerateType   string
	GenerateParams interface{}
	GenerateResult interface{}
	GenerateErr    error
}

var _ bconfigserver.Client = &FakeConfigServerClient{}

func (c *FakeConfigServerClient) Generate(name, type_ string, params interface{}) (interface{}, error) {
	c.GenerateName = name
	c.GenerateType = type_
	c.GenerateParams = params
	return c.GenerateResult, c.GenerateErr
}
