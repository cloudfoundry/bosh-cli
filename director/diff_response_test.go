package director

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Deployment", func() {
	Describe("Diff", func() {
		Describe("ConvertDiffResponseToDiff", func() {
			It("converts", func() {
				diffResponse := DeploymentDiffResponse{
					Context: map[string]interface{}{
						"cloud_config_id":   2,
						"runtime_config_id": nil,
					},
					Diff: [][]interface{}{
						[]interface{}{"name: simple manifest", nil},
						[]interface{}{"properties:", nil},
						[]interface{}{"  - property1", "removed"},
						[]interface{}{"  - property2", "added"},
					},
				}
				diff := ConvertDiffResponseToDiff(diffResponse)

				Expect(len(diff.Diff)).To(Equal(len(diffResponse.Diff)))
				Expect(diff.Diff[0]).To(Equal([]interface{}{"name: simple manifest", nil}))
				Expect(diff.Diff[1]).To(Equal([]interface{}{"properties:", nil}))
				Expect(diff.Diff[2]).To(Equal([]interface{}{"  - property1", "removed"}))
				Expect(diff.Diff[3]).To(Equal([]interface{}{"  - property2", "added"}))

				Expect(diff.context).To(Equal(map[string]interface{}{
					"cloud_config_id":   2,
					"runtime_config_id": nil,
				}))
			})
		})
	})
})
