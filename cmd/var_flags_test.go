package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	. "github.com/cloudfoundry/bosh-init/director/template"
)

var _ = Describe("VarFlags", func() {
	Describe("AsVariables", func() {
		It("merges files and then kvs", func() {
			flags := VarFlags{
				VarKVs: []VarKV{
					{Name: "name1", Value: "val1"},
					{Name: "name2", Value: "val2"},
					{Name: "name2", Value: "val2-over"},
					{Name: "name5", Value: "val5-over"},
				},
				VarsFiles: []VarsFileArg{
					{Vars: Variables{
						"name3": "val3",
						"name4": "val4",
					}},
					{Vars: Variables{
						"name5": "val5",
						"name6": "val6",
						"name4": "val4-over",
					}},
				},
			}

			vars := flags.AsVariables()
			Expect(vars).To(Equal(Variables{
				"name1": "val1",
				"name2": "val2-over",
				"name3": "val3",
				"name4": "val4-over",
				"name5": "val5-over",
				"name6": "val6",
			}))
		})
	})
})
