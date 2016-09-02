package cmd_test

import (
	"errors"

	semver "github.com/cppforlife/go-semi-semantic/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	fakecmd "github.com/cloudfoundry/bosh-cli/cmd/fakes"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	fakedir "github.com/cloudfoundry/bosh-cli/director/fakes"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
)

var _ = Describe("DeployCmd", func() {
	var (
		ui               *fakeui.FakeUI
		deployment       *fakedir.FakeDeployment
		uploadReleaseCmd *fakecmd.FakeReleaseUploadingCmd
		command          DeployCmd
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		deployment = &fakedir.FakeDeployment{
			NameStub: func() string { return "dep" },
		}
		uploadReleaseCmd = &fakecmd.FakeReleaseUploadingCmd{}
		command = NewDeployCmd(ui, deployment, uploadReleaseCmd)
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

			bytes, updateOpts := deployment.UpdateArgsForCall(0)
			Expect(bytes).To(Equal([]byte("name: dep\n")))
			Expect(updateOpts).To(Equal(boshdir.UpdateOpts{}))
		})

		It("deploys manifest allowing to recreate, fix, and skip drain", func() {
			opts.Recreate = true
			opts.Fix = true
			opts.SkipDrain = boshdir.SkipDrain{All: true}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.UpdateCallCount()).To(Equal(1))

			bytes, updateOpts := deployment.UpdateArgsForCall(0)
			Expect(bytes).To(Equal([]byte("name: dep\n")))
			Expect(updateOpts).To(Equal(boshdir.UpdateOpts{
				Recreate:  true,
				Fix:       true,
				SkipDrain: boshdir.SkipDrain{All: true},
			}))
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

			bytes, _ := deployment.UpdateArgsForCall(0)
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

		It("does not upload releases and deploy if confirmation is rejected", func() {
			opts.Args.Manifest = FileBytesArg{
				Bytes: []byte(`
name: dep
releases:
- name: capi
  sha1: capi-sha1
  url: https://capi-url
  version: 1+capi
`),
			}

			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(uploadReleaseCmd.RunCallCount()).To(Equal(0))
			Expect(deployment.UpdateCallCount()).To(Equal(0))
		})

		It("returns an error if diffing failed", func() {
			deployment.DiffReturns(boshdir.DiffLines{}, errors.New("Fetching diff result"))

			err := act()
			Expect(err).To(HaveOccurred())
		})

		It("gets the diff from the deployment", func() {
			expectedDiff := boshdir.DiffLines{
				[]interface{}{
					"some line that stayed", nil,
				}, []interface{}{
					"some line that was added", "added",
				}, []interface{}{
					"some line that was removed", "removed",
				},
			}

			deployment.DiffReturns(expectedDiff, nil)
			err := act()
			Expect(err).ToNot(HaveOccurred())
			Expect(deployment.DiffCallCount()).To(Equal(1))
			Expect(ui.Said).To(ContainElement("  some line that stayed\n"))
			Expect(ui.Said).To(ContainElement("+ some line that was added\n"))
			Expect(ui.Said).To(ContainElement("- some line that was removed\n"))
		})

		It("uploads remote releases skipping releases without url", func() {
			opts.Args.Manifest = FileBytesArg{
				Bytes: []byte(`
name: dep
releases:
- name: capi
  sha1: capi-sha1
  url: https://capi-url
  version: 1+capi
- name: rel-without-upload
  version: 1+rel
- name: consul
  sha1: consul-sha1
  url: https://consul-url
  version: 1+consul
`),
			}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(uploadReleaseCmd.RunCallCount()).To(Equal(2))

			Expect(uploadReleaseCmd.RunArgsForCall(0)).To(Equal(UploadReleaseOpts{
				Name:    "capi",
				Args:    UploadReleaseArgs{URL: URLArg("https://capi-url")},
				SHA1:    "capi-sha1",
				Version: VersionArg(semver.MustNewVersionFromString("1+capi")),
			}))

			Expect(uploadReleaseCmd.RunArgsForCall(1)).To(Equal(UploadReleaseOpts{
				Name:    "consul",
				Args:    UploadReleaseArgs{URL: URLArg("https://consul-url")},
				SHA1:    "consul-sha1",
				Version: VersionArg(semver.MustNewVersionFromString("1+consul")),
			}))
		})

		It("returns error and does not deploy if uploading release fails", func() {
			opts.Args.Manifest = FileBytesArg{
				Bytes: []byte(`
name: dep
releases:
- name: capi
  sha1: capi-sha1
  url: https://capi-url
  version: 1+capi
`),
			}
			uploadReleaseCmd.RunReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(deployment.UpdateCallCount()).To(Equal(0))
		})

		It("returns an error if release version cannot be parsed", func() {
			opts.Args.Manifest = FileBytesArg{
				Bytes: []byte(`
name: dep
releases:
- name: capi
  sha1: capi-sha1
  url: https://capi-url
  version: 1+capi+capi
`),
			}

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected version '1+capi+capi' to match version format"))

			Expect(uploadReleaseCmd.RunCallCount()).To(Equal(0))
			Expect(deployment.UpdateCallCount()).To(Equal(0))
		})

		It("returns an error if release version cannot be parsed", func() {
			opts.Args.Manifest = FileBytesArg{
				Bytes: []byte(`
name: dep
releases:
- name: capi
  sha1: capi-sha1
  url: https://capi-url
  version: 1+capi+capi
`),
			}

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Expected version '1+capi+capi' to match version format"))

			Expect(uploadReleaseCmd.RunCallCount()).To(Equal(0))
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
