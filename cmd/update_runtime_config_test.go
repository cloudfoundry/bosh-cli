package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	fakedir "github.com/cloudfoundry/bosh-init/director/fakes"
	boshtpl "github.com/cloudfoundry/bosh-init/director/template"
	fakeui "github.com/cloudfoundry/bosh-init/ui/fakes"
)

var _ = Describe("UpdateRuntimeConfigCmd", func() {
	var (
		ui       *fakeui.FakeUI
		director *fakedir.FakeDirector
		command  UpdateRuntimeConfigCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		director = &fakedir.FakeDirector{}
		command = NewUpdateRuntimeConfigCmd(ui, director)
	})

	Describe("Run", func() {
		var (
			opts UpdateRuntimeConfigOpts
		)

		BeforeEach(func() {
			opts = UpdateRuntimeConfigOpts{
				Args: UpdateRuntimeConfigArgs{
					RuntimeConfig: FileBytesArg{Bytes: []byte("runtime-config")},
				},
			}
		})

		act := func() error { return command.Run(opts) }

		It("updates runtime config", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.UpdateRuntimeConfigCallCount()).To(Equal(1))

			bytes := director.UpdateRuntimeConfigArgsForCall(0)
			Expect(bytes).To(Equal([]byte("runtime-config")))
		})

		It("updates runtime config with evaluated vars", func() {
			opts.Args.RuntimeConfig = FileBytesArg{
				Bytes: []byte("name1: ((name1))\nname2: ((name2))"),
			}

			opts.VarKVs = []boshtpl.VarKV{
				{Name: "name1", Value: "val1-from-kv"},
			}

			opts.VarsFiles = []boshtpl.VarsFileArg{
				{Vars: boshtpl.Variables(map[string]string{"name1": "val1-from-file"})},
				{Vars: boshtpl.Variables(map[string]string{"name2": "val2-from-file"})},
			}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(director.UpdateRuntimeConfigCallCount()).To(Equal(1))

			bytes := director.UpdateRuntimeConfigArgsForCall(0)
			Expect(bytes).To(Equal([]byte("name1: \"val1-from-kv\"\nname2: \"val2-from-file\"")))
		})

		It("does not stop if confirmation is rejected", func() {
			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(director.UpdateRuntimeConfigCallCount()).To(Equal(0))
		})

		It("returns error if updating failed", func() {
			director.UpdateRuntimeConfigReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
