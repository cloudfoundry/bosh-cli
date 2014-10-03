package deployment_test

import (
	// "errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/deployment"
)

var _ = Describe("Deployment", func() {
	Describe("NetworksSpec", func() {
		var (
			deployment Deployment
		)

		Context("when the deployment has networks", func() {
			BeforeEach(func() {
				deployment = Deployment{
					Networks: []Network{
						{
							Name: "fake-network-name",
							Type: "dynamic",
						},
						{
							Name: "fake-network-name-2",
							Type: "dynamic",
						},
					},
				}
			})

			It("is a map of the networks in spec form", func() {
				Expect(deployment.NetworksSpec()).To(Equal(map[string]interface{}{
					"fake-network-name": map[string]interface{}{
						"type":             "dynamic",
						"cloud_properties": map[string]interface{}{},
					},
					"fake-network-name-2": map[string]interface{}{
						"type":             "dynamic",
						"cloud_properties": map[string]interface{}{},
					},
				}))
			})
		})

		Context("when the deployment does not have networks", func() {
			BeforeEach(func() {
				deployment = Deployment{
					Networks: []Network{},
				}
			})

			It("is an empty map", func() {
				Expect(deployment.NetworksSpec()).To(Equal(map[string]interface{}{}))
			})
		})
	})
})
