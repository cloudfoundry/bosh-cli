package cmd_test

import (
	"errors"

	"github.com/cppforlife/go-patch/patch"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	fakedir "github.com/cloudfoundry/bosh-cli/director/directorfakes"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
)

var _ = Describe("UpdateConfigCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  UpdateConfigCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = NewUpdateConfigCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			opts UpdateConfigOpts
		)

		BeforeEach(func() {
			opts = UpdateConfigOpts{
				Args: UpdateConfigArgs{
					Type:   "my-type",
					Config: FileBytesArg{Bytes: []byte("fake-config")},
				},
				Name: "my-name",
			}
		})

		act := func() error { return command.Run(opts) }

		It("uploads new config", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.UpdateConfigCallCount()).To(Equal(1))

			t, name, bytes := director.UpdateConfigArgsForCall(0)
			Expect(t).To(Equal("my-type"))
			Expect(name).To(Equal("my-name"))
			Expect(bytes).To(Equal([]byte("fake-config\n")))
		})

		It("updates templated config", func() {
			opts.Args.Config = FileBytesArg{
				Bytes: []byte("name1: ((name1))\nname2: ((name2))"),
			}

			opts.VarKVs = []boshtpl.VarKV{
				{Name: "name1", Value: "val1-from-kv"},
			}

			opts.VarsFiles = []boshtpl.VarsFileArg{
				{Vars: boshtpl.StaticVariables(map[string]interface{}{"name1": "val1-from-file"})},
				{Vars: boshtpl.StaticVariables(map[string]interface{}{"name2": "val2-from-file"})},
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

			Expect(director.UpdateConfigCallCount()).To(Equal(1))

			t, name, bytes := director.UpdateConfigArgsForCall(0)
			Expect(t).To(Equal("my-type"))
			Expect(name).To(Equal("my-name"))
			Expect(bytes).To(Equal([]byte("name1: val1-from-kv\nname2: val2-from-file\nxyz: val\n")))
		})

		It("does not stop if confirmation is rejected", func() {
			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(director.UpdateConfigCallCount()).To(Equal(0))
		})

		It("returns error if updating failed", func() {
			director.UpdateConfigReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		Context("when uploading an empty YAML document", func() {
			BeforeEach(func() {
				opts = UpdateConfigOpts{
					Args: UpdateConfigArgs{
						Type:   "my-type",
						Config: FileBytesArg{Bytes: []byte("---")},
					},
					Name: "",
				}
			})

			It("returns YAML null", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())
				_, _, bytes := director.UpdateConfigArgsForCall(0)
				Expect(bytes).To(Equal([]byte("null\n")))
			})
		})
	})
})
