package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	. "github.com/cloudfoundry/bosh-cli/director/template"
)

var _ = Describe("VarFlags", func() {
	Describe("AsVariables", func() {
		It("merges files and then kvs", func() {
			flags := VarFlags{
				VarKVs: []VarKV{
					{Name: "kv", Value: "kv"},
					{Name: "kv_precedence", Value: "kv1"},
					{Name: "kv_precedence", Value: "kv2"},
					{Name: "kv_file_precedence", Value: "kv"},
				},
				VarsFiles: []VarsFileArg{
					{Vars: Variables{
						"file": "file",
						"file_precedence": "file",
					}},
					{Vars: Variables{
						"kv_file_precedence": "file2",
						"file2": "file2",
						"file_precedence": "file2",
					}},
				},
			}

			vars := flags.AsVariables()
			Expect(vars).To(Equal(Variables{
				"kv": "kv",
				"kv_precedence": "kv2",
				"file": "file",
				"file_precedence": "file2",
				"kv_file_precedence": "kv",
				"file2": "file2",
			}))
		})
	})
})
