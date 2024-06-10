package cmd_test

import (
	"errors"

	"github.com/cppforlife/go-patch/patch"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"
)

var _ = Describe("UpdateConfigCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  cmd.UpdateConfigCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewUpdateConfigCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			updateConfigOpts opts.UpdateConfigOpts
		)

		BeforeEach(func() {
			updateConfigOpts = opts.UpdateConfigOpts{
				Args: opts.UpdateConfigArgs{
					Config: opts.FileBytesArg{Bytes: []byte("fake-config")},
				},
				Type: "my-type",
				Name: "my-name",
			}
		})

		act := func() error { return command.Run(updateConfigOpts) }

		It("uploads new config", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.UpdateConfigCallCount()).To(Equal(1))

			t, name, expectedLatestId, bytes := director.UpdateConfigArgsForCall(0)
			Expect(t).To(Equal("my-type"))
			Expect(name).To(Equal("my-name"))
			Expect(bytes).To(Equal([]byte("fake-config\n")))
			Expect(expectedLatestId).To(Equal(""))
		})

		It("updates templated config", func() {
			updateConfigOpts.Args.Config = opts.FileBytesArg{
				Bytes: []byte("name1: ((name1))\nname2: ((name2))"),
			}

			updateConfigOpts.VarKVs = []boshtpl.VarKV{
				{Name: "name1", Value: "val1-from-kv"},
			}

			updateConfigOpts.VarsFiles = []boshtpl.VarsFileArg{
				{Vars: boshtpl.StaticVariables(map[string]interface{}{"name1": "val1-from-file"})},
				{Vars: boshtpl.StaticVariables(map[string]interface{}{"name2": "val2-from-file"})},
			}

			updateConfigOpts.OpsFiles = []opts.OpsFileArg{
				{
					Ops: patch.Ops([]patch.Op{
						patch.ReplaceOp{Path: patch.MustNewPointerFromString("/xyz?"), Value: "val"},
					}),
				},
			}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.UpdateConfigCallCount()).To(Equal(1))

			t, name, _, bytes := director.UpdateConfigArgsForCall(0)
			Expect(t).To(Equal("my-type"))
			Expect(name).To(Equal("my-name"))
			Expect(bytes).To(Equal([]byte("name1: val1-from-kv\nname2: val2-from-file\nxyz: val\n")))
		})

		It("outputs a table that should be transposed", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(ui.Table.Transpose).To(Equal(true))
		})

		It("output table contains headers and rows", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(ui.Table.Header).To(Equal([]boshtbl.Header{
				boshtbl.NewHeader("ID"),
				boshtbl.NewHeader("Type"),
				boshtbl.NewHeader("Name"),
				boshtbl.NewHeader("Created At"),
				boshtbl.NewHeader("Content"),
			}))
			Expect(ui.Table.Rows).To(Equal([][]boshtbl.Value{
				{
					boshtbl.NewValueString(""),
					boshtbl.NewValueString(""),
					boshtbl.NewValueString(""),
					boshtbl.NewValueString(""),
					boshtbl.NewValueString(""),
				},
			}))
		})

		It("does not update if confirmation is rejected", func() {
			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(director.UpdateConfigCallCount()).To(Equal(0))
		})

		It("returns error if updating failed", func() {
			director.UpdateConfigReturns(boshdir.Config{}, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns an error if diffing failed", func() {
			director.DiffConfigReturns(boshdir.ConfigDiff{}, errors.New("Fetching diff result"))

			err := act()
			Expect(err).To(HaveOccurred())
		})

		It("gets the diff from the config", func() {
			diff := [][]interface{}{
				[]interface{}{"some line that stayed", nil},
				[]interface{}{"some line that was added", "added"},
				[]interface{}{"some line that was removed", "removed"},
			}

			expectedDiff := boshdir.NewConfigDiff(diff)
			director.DiffConfigReturns(expectedDiff, nil)
			err := act()
			Expect(err).ToNot(HaveOccurred())
			Expect(director.DiffConfigCallCount()).To(Equal(1))
			Expect(ui.Said).To(ContainElement("  some line that stayed\n"))
			Expect(ui.Said).To(ContainElement("+ some line that was added\n"))
			Expect(ui.Said).To(ContainElement("- some line that was removed\n"))
		})

		Context("when expected-latest-id is specified", func() {
			BeforeEach(func() {
				updateConfigOpts = opts.UpdateConfigOpts{
					Args: opts.UpdateConfigArgs{
						Config: opts.FileBytesArg{Bytes: []byte("---")},
					},
					Type:             "my-type",
					Name:             "my-name",
					ExpectedLatestId: "123",
				}
			})

			It("passes expected latest id when calling update config", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())
				_, _, expectedLatestId, _ := director.UpdateConfigArgsForCall(0)
				Expect(expectedLatestId).To(Equal("123"))
			})
		})

		Context("when expected-latest-id is not specified", func() {
			Context("when a config is already uploaded", func() {
				It("calls update config with the latest id returned by diff config", func() {
					director.DiffConfigReturns(boshdir.ConfigDiff{Diff: [][]interface{}{}, FromId: "1"}, nil)

					err := act()
					Expect(err).ToNot(HaveOccurred())
					_, _, expectedLatestId, _ := director.UpdateConfigArgsForCall(0)
					Expect(expectedLatestId).To(Equal("1"))
				})
			})
			Context("when no config is uploaded", func() {
				It("calls update config without a latest id", func() {
					err := act()
					Expect(err).ToNot(HaveOccurred())
					_, _, expectedLatestId, _ := director.UpdateConfigArgsForCall(0)
					Expect(expectedLatestId).To(Equal(""))
				})
			})
		})

		Context("when uploading an empty YAML document", func() {
			BeforeEach(func() {
				updateConfigOpts = opts.UpdateConfigOpts{
					Args: opts.UpdateConfigArgs{
						Config: opts.FileBytesArg{Bytes: []byte("---")},
					},
					Type: "my-type",
					Name: "",
				}
			})

			It("returns YAML null", func() {
				err := act()
				Expect(err).ToNot(HaveOccurred())
				_, _, _, bytes := director.UpdateConfigArgsForCall(0)
				Expect(bytes).To(Equal([]byte("null\n")))
			})
		})
	})
})
