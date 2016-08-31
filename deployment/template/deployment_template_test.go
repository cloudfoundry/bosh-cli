package template_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/deployment/template"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
)

var _ = Describe("DeploymentTemplate", func() {
	It("can template values into a struct with byte slice", func() {
		deploymentTemplate := NewDeploymentTemplate([]byte("((key))"))
		variables := boshtpl.Variables{"key": "foo"}

		result, err := deploymentTemplate.Evaluate(variables)
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Content()).To(Equal([]byte("foo\n")))
	})

	It("returns a struct that can return the SHA2_512 of the struct", func() {
		deploymentTemplate := NewDeploymentTemplate([]byte("((key))"))
		variables := boshtpl.Variables{"key": "foo"}

		result, err := deploymentTemplate.Evaluate(variables)
		Expect(err).NotTo(HaveOccurred())
		asString := result.SHA()
		Expect(asString).To(Equal("0cf9180a764aba863a67b6d72f0918bc131c6772642cb2dce5a34f0a702f9470ddc2bf125c12198b1995c233c34b4afd346c54a2334c350a948a51b6e8b4e6b6"))
	})
})
