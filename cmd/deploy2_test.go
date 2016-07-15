package cmd_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/cmd"
	boshdir "github.com/cloudfoundry/bosh-init/director"
	fakedir "github.com/cloudfoundry/bosh-init/director/fakes"
	boshtpl "github.com/cloudfoundry/bosh-init/director/template"
	fakeui "github.com/cloudfoundry/bosh-init/ui/fakes"
)

var _ = Describe("Deploy2Cmd", func() {
	var (
		ui         *fakeui.FakeUI
		deployment *fakedir.FakeDeployment
		command    Deploy2Cmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		deployment = &fakedir.FakeDeployment{
			NameStub: func() string { return "dep" },
		}
		command = NewDeploy2Cmd(ui, deployment)
	})

	Describe("Run", func() {
		var (
			opts DeployOpts
		)

		BeforeEach(func() {
			opts = DeployOpts{
				Args: DeployArgs{
					Manifest: FileBytesArg{Bytes: []byte("name: dep")},
				},
			}
		})

		act := func() error { return command.Run(opts) }

		It("deploys manifest", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.UpdateCallCount()).To(Equal(1))

			bytes, recreate, sd := deployment.UpdateArgsForCall(0)
			Expect(bytes).To(Equal([]byte("name: dep\n")))
			Expect(recreate).To(BeFalse())
			Expect(sd).To(Equal(boshdir.SkipDrain{}))
		})

		It("deploys manifest allowing to recreate", func() {
			opts.Recreate = true

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.UpdateCallCount()).To(Equal(1))

			bytes, recreate, sd := deployment.UpdateArgsForCall(0)
			Expect(bytes).To(Equal([]byte("name: dep\n")))
			Expect(recreate).To(BeTrue())
			Expect(sd).To(Equal(boshdir.SkipDrain{}))
		})

		It("deploys manifest allowing to skip drain scripts", func() {
			opts.SkipDrain = boshdir.SkipDrain{All: true}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.UpdateCallCount()).To(Equal(1))

			bytes, recreate, sd := deployment.UpdateArgsForCall(0)
			Expect(bytes).To(Equal([]byte("name: dep\n")))
			Expect(recreate).To(BeFalse())
			Expect(sd).To(Equal(boshdir.SkipDrain{All: true}))
		})

		It("deploys manifest with evaluated vars", func() {
			opts.Args.Manifest = FileBytesArg{
				Bytes: []byte("name: dep\nname1: ((name1))\nname2: ((name2))\n"),
			}

			opts.VarKVs = []boshtpl.VarKV{
				{Name: "name1", Value: "val1-from-kv"},
			}

			opts.VarsFiles = []boshtpl.VarsFileArg{
				{Vars: boshtpl.Variables(map[string]interface{}{"name1": "val1-from-file"})},
				{Vars: boshtpl.Variables(map[string]interface{}{"name2": "val2-from-file"})},
			}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.UpdateCallCount()).To(Equal(1))

			bytes, _, _ := deployment.UpdateArgsForCall(0)
			Expect(bytes).To(Equal([]byte("name: dep\nname1: val1-from-kv\nname2: val2-from-file\n")))
		})

		It("does not deploy if name specified in the manifest does not match deployment's name", func() {
			opts.Args.Manifest = FileBytesArg{
				Bytes: []byte("name: other-name"),
			}

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected manifest to specify deployment name 'dep' but was 'other-name'"))

			Expect(deployment.UpdateCallCount()).To(Equal(0))
		})

		It("does not deploy if confirmation is rejected", func() {
			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(deployment.UpdateCallCount()).To(Equal(0))
		})

		It("returns error if deploying failed", func() {
			deployment.UpdateReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})
})
