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
)

var _ = Describe("UpdateCloudConfigCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  cmd.UpdateCloudConfigCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = cmd.NewUpdateCloudConfigCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			updateCloudConfigOpts opts.UpdateCloudConfigOpts
		)

		BeforeEach(func() {
			updateCloudConfigOpts = opts.UpdateCloudConfigOpts{
				Args: opts.UpdateCloudConfigArgs{
					CloudConfig: opts.FileBytesArg{Bytes: []byte("cloud-config")},
				},
				Name: "angry-smurf",
			}
		})

		act := func() error { return command.Run(updateCloudConfigOpts) }

		It("updates cloud config", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.UpdateCloudConfigCallCount()).To(Equal(1))

			name, bytes := director.UpdateCloudConfigArgsForCall(0)
			Expect(name).To(Equal("angry-smurf"))
			Expect(bytes).To(Equal([]byte("cloud-config\n")))
		})

		It("updates templated cloud config", func() {
			updateCloudConfigOpts.Args.CloudConfig = opts.FileBytesArg{
				Bytes: []byte("name1: ((name1))\nname2: ((name2))"),
			}

			updateCloudConfigOpts.VarKVs = []boshtpl.VarKV{
				{Name: "name1", Value: "val1-from-kv"},
			}

			updateCloudConfigOpts.VarsFiles = []boshtpl.VarsFileArg{
				{Vars: boshtpl.StaticVariables(map[string]interface{}{"name1": "val1-from-file"})},
				{Vars: boshtpl.StaticVariables(map[string]interface{}{"name2": "val2-from-file"})},
			}

			updateCloudConfigOpts.OpsFiles = []opts.OpsFileArg{
				{
					Ops: patch.Ops([]patch.Op{
						patch.ReplaceOp{Path: patch.MustNewPointerFromString("/xyz?"), Value: "val"},
					}),
				},
			}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.UpdateCloudConfigCallCount()).To(Equal(1))

			name, bytes := director.UpdateCloudConfigArgsForCall(0)
			Expect(name).To(Equal("angry-smurf"))
			Expect(bytes).To(Equal([]byte("name1: val1-from-kv\nname2: val2-from-file\nxyz: val\n")))
		})

		It("returns an error if diffing failed", func() {
			director.DiffCloudConfigReturns(boshdir.ConfigDiff{}, errors.New("Fetching diff result"))

			err := act()
			Expect(err).To(HaveOccurred())
		})

		It("gets the diff from the deployment", func() {
			diff := [][]interface{}{
				[]interface{}{"some line that stayed", nil},
				[]interface{}{"some line that was added", "added"},
				[]interface{}{"some line that was removed", "removed"},
			}

			expectedDiff := boshdir.NewConfigDiff(diff)
			director.DiffCloudConfigReturns(expectedDiff, nil)
			err := act()
			Expect(err).ToNot(HaveOccurred())
			Expect(director.DiffCloudConfigCallCount()).To(Equal(1))
			Expect(ui.Said).To(ContainElement("  some line that stayed\n"))
			Expect(ui.Said).To(ContainElement("+ some line that was added\n"))
			Expect(ui.Said).To(ContainElement("- some line that was removed\n"))
		})

		It("does not stop if confirmation is rejected", func() {
			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(director.UpdateCloudConfigCallCount()).To(Equal(0))
		})

		It("returns error if updating failed", func() {
			director.UpdateCloudConfigReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
