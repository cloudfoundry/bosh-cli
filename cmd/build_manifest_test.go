package cmd_test

import (
	"github.com/cppforlife/go-patch/patch"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/uifakes"
)

var _ = Describe("BuildManifestCmd", func() {
	var (
		ui      *fakeui.FakeUI
		command BuildManifestCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		command = NewBuildManifestCmd(ui)
	})

	Describe("Run", func() {
		var (
			opts BuildManifestOpts
		)

		BeforeEach(func() {
			opts = BuildManifestOpts{}
		})

		act := func() error { return command.Run(opts) }

		It("shows templated manifest", func() {
			opts.Args.Manifest = FileBytesArg{
				Bytes: []byte("name1: ((name1))\nname2: ((name2))"),
			}

			opts.VarKVs = []boshtpl.VarKV{
				{Name: "name1", Value: "val1-from-kv"},
			}

			opts.VarsFiles = []boshtpl.VarsFileArg{
				{Vars: boshtpl.Variables(map[string]interface{}{"name1": "val1-from-file"})},
				{Vars: boshtpl.Variables(map[string]interface{}{"name2": "val2-from-file"})},
			}

			opts.OpsFiles = []OpsFileArg{
				{
					Ops: patch.Ops([]patch.Op{
						patch.ReplaceOp{Path: patch.MustNewPointerFromString("/xyz?"), Value: "val"},
					}),
				},
			}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			bytes := "name1: val1-from-kv\nname2: val2-from-file\nxyz: val\n"
			Expect(ui.Blocks).To(Equal([]string{bytes}))
		})

		It("returns error if variables are not found in templated manifest if var-errors flag is set", func() {
			opts.Args.Manifest = FileBytesArg{
				Bytes: []byte("name1: ((name1))\nname2: ((name2))"),
			}

			opts.VarKVs = []boshtpl.VarKV{
				{Name: "name1", Value: "val1-from-kv"},
			}

			opts.VarErrors = true

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected to find variables: name2"))
		})
	})
})
