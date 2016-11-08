package cmd_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	. "github.com/cloudfoundry/bosh-cli/director/template"
)

var _ = Describe("VarFlags", func() {
	Describe("AsVariables", func() {
		It("merges environment variables, then files and then kvs", func() {
			flags := VarFlags{
				VarKVs: []VarKV{
					{Name: "kv", Value: "kv"},
					{Name: "kv_precedence", Value: "kv1"},
					{Name: "kv_precedence", Value: "kv2"},
					{Name: "kv_env_precedence", Value: "kv"},
					{Name: "kv_file_precedence", Value: "kv"},
					{Name: "kv_file_env_precedence", Value: "kv"},
				},
				VarsFiles: []VarsFileArg{
					{Vars: Variables{
						"file":            "file",
						"file_precedence": "file",
					}},
					{Vars: Variables{
						"file_env_precedence":    "file2",
						"kv_file_env_precedence": "file2",
						"kv_file_precedence":     "file2",
						"file2":                  "file2",
						"file_precedence":        "file2",
					}},
				},
				VarsEnvs: []VarsEnvArg{
					{Vars: Variables{
						"env":            "env",
						"env_precedence": "env",
					}},
					{Vars: Variables{
						"kv_env_precedence":      "env2",
						"file_env_precedence":    "env2",
						"kv_file_env_precedence": "env2",
						"env2":           "env2",
						"env_precedence": "env2",
					}},
				},
			}

			vars := flags.AsVariables()
			Expect(vars).To(Equal(Variables{
				"kv":                     "kv",
				"kv_precedence":          "kv2",
				"file":                   "file",
				"file_precedence":        "file2",
				"kv_file_precedence":     "kv",
				"file2":                  "file2",
				"env2":                   "env2",
				"env":                    "env",
				"env_precedence":         "env2",
				"kv_file_env_precedence": "kv",
				"file_env_precedence":    "file2",
				"kv_env_precedence":      "kv",
			}))
		})
	})
})
