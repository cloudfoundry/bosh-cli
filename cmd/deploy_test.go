package cmd_test

import (
	"errors"
	"github.com/cppforlife/go-patch/patch"
	"github.com/cppforlife/go-semi-semantic/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	fakecmd "github.com/cloudfoundry/bosh-cli/v7/cmd/cmdfakes"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	fakedir "github.com/cloudfoundry/bosh-cli/v7/director/directorfakes"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("DeployCmd", func() {
	var (
		ui              *fakeui.FakeUI
		deployment      *fakedir.FakeDeployment
		releaseUploader *fakecmd.FakeReleaseUploader
		director        *fakedir.FakeDirector
		command         cmd.DeployCmd
		release         *fakedir.FakeRelease
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		deployment = &fakedir.FakeDeployment{
			NameStub: func() string { return "dep" },
		}

		releaseUploader = &fakecmd.FakeReleaseUploader{
			UploadReleasesStub: func(bytes []byte) ([]byte, error) { return bytes, nil },
		}

		director = &fakedir.FakeDirector{}

		command = cmd.NewDeployCmd(ui, deployment, releaseUploader, director)

		release = &fakedir.FakeRelease{
			NameStub: func() string { return "ReleaseName" },
			VersionStub: func() version.Version {
				return version.MustNewVersionFromString("1")
			},
		}
	})

	Describe("Run", func() {
		var (
			deployOpts opts.DeployOpts
		)

		BeforeEach(func() {
			deployOpts = opts.DeployOpts{
				Args: opts.DeployArgs{
					Manifest: opts.FileBytesArg{Bytes: []byte("name: dep")},
				},
			}
		})

		act := func() error { return command.Run(deployOpts) }

		It("deploys manifest", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.UpdateCallCount()).To(Equal(1))

			bytes, updateOpts := deployment.UpdateArgsForCall(0)
			Expect(bytes).To(Equal([]byte("name: dep\n")))
			Expect(updateOpts).To(Equal(boshdir.UpdateOpts{}))
		})

		It("deploys manifest allowing to recreate, recreate persistent disks, fix, and skip drain", func() {
			deployOpts.RecreatePersistentDisks = true
			deployOpts.Recreate = true
			deployOpts.Fix = true
			deployOpts.SkipDrain = boshdir.SkipDrains{boshdir.SkipDrain{All: true}}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.UpdateCallCount()).To(Equal(1))

			bytes, updateOpts := deployment.UpdateArgsForCall(0)
			Expect(bytes).To(Equal([]byte("name: dep\n")))
			Expect(updateOpts).To(Equal(boshdir.UpdateOpts{
				RecreatePersistentDisks: true,
				Recreate:                true,
				Fix:                     true,
				SkipDrain:               boshdir.SkipDrains{boshdir.SkipDrain{All: true}},
			}))
		})

		It("deploys manifest allowing to dry_run", func() {
			deployOpts.DryRun = true

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.UpdateCallCount()).To(Equal(1))

			bytes, updateOpts := deployment.UpdateArgsForCall(0)
			Expect(bytes).To(Equal([]byte("name: dep\n")))
			Expect(updateOpts).To(Equal(boshdir.UpdateOpts{
				DryRun: true,
			}))
		})

		It("deploys manifest allowing to force latest variables", func() {
			deployOpts.ForceLatestVariables = true

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.UpdateCallCount()).To(Equal(1))

			bytes, updateOpts := deployment.UpdateArgsForCall(0)
			Expect(bytes).To(Equal([]byte("name: dep\n")))
			Expect(updateOpts).To(Equal(boshdir.UpdateOpts{
				ForceLatestVariables: true,
			}))
		})

		It("deploys templated manifest", func() {
			deployOpts.Args.Manifest = opts.FileBytesArg{
				Bytes: []byte("name: dep\nname1: ((name1))\nname2: ((name2))\n"),
			}

			deployOpts.VarKVs = []boshtpl.VarKV{
				{Name: "name1", Value: "val1-from-kv"},
			}

			deployOpts.VarsFiles = []boshtpl.VarsFileArg{
				{Vars: boshtpl.StaticVariables(map[string]interface{}{"name1": "val1-from-file"})},
				{Vars: boshtpl.StaticVariables(map[string]interface{}{"name2": "val2-from-file"})},
			}

			deployOpts.OpsFiles = []opts.OpsFileArg{
				{
					Ops: patch.Ops([]patch.Op{
						patch.ReplaceOp{Path: patch.MustNewPointerFromString("/xyz?"), Value: "val"},
					}),
				},
			}

			err := act()
			Expect(err).ToNot(HaveOccurred())

			Expect(deployment.UpdateCallCount()).To(Equal(1))

			bytes, _ := deployment.UpdateArgsForCall(0)
			Expect(bytes).To(Equal([]byte("name: dep\nname1: val1-from-kv\nname2: val2-from-file\nxyz: val\n")))
		})

		It("does not deploy if name specified in the manifest does not match deployment's name", func() {
			deployOpts.Args.Manifest = opts.FileBytesArg{
				Bytes: []byte("name: other-name"),
			}

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(
				"Expected manifest to specify deployment name 'dep' but was 'other-name'"))

			Expect(deployment.UpdateCallCount()).To(Equal(0))
		})

		It("uploads releases provided in the manifest after manifest has been interpolated", func() {
			deployOpts.Args.Manifest = opts.FileBytesArg{
				Bytes: []byte("name: dep\nbefore-upload-manifest: ((key))"),
			}

			deployOpts.VarKVs = []boshtpl.VarKV{
				{Name: "key", Value: "key-val"},
			}

			releaseUploader.UploadReleasesReturns([]byte("after-upload-manifest"), nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			bytes := releaseUploader.UploadReleasesArgsForCall(0)
			Expect(bytes).To(Equal([]byte("before-upload-manifest: key-val\nname: dep\n")))

			Expect(deployment.UpdateCallCount()).To(Equal(1))

			bytes, _ = deployment.UpdateArgsForCall(0)
			Expect(bytes).To(Equal([]byte("after-upload-manifest")))
		})

		It("uploads releases provided in the manifest with fix after manifest has been interpolated", func() {
			deployOpts.Args.Manifest = opts.FileBytesArg{
				Bytes: []byte("name: dep\nbefore-upload-manifest-with-fix: ((key))"),
			}

			deployOpts.VarKVs = []boshtpl.VarKV{
				{Name: "key", Value: "key-val"},
			}

			deployOpts.FixReleases = true

			releaseUploader.UploadReleasesWithFixReturns([]byte("after-upload-manifest-with-fix"), nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())

			bytes := releaseUploader.UploadReleasesWithFixArgsForCall(0)
			Expect(bytes).To(Equal([]byte("before-upload-manifest-with-fix: key-val\nname: dep\n")))

			Expect(deployment.UpdateCallCount()).To(Equal(1))

			bytes, _ = deployment.UpdateArgsForCall(0)
			Expect(bytes).To(Equal([]byte("after-upload-manifest-with-fix")))
		})

		It("skips the upload of all releases in the corresponding deployment", func() {
			var releases = []boshdir.Release{release}

			deployment.ReleasesReturns(releases, nil)

			deployOpts.SkipDownloadReleases = true

			err := act()
			Expect(err).ToNot(HaveOccurred())
			Expect(releaseUploader.UploadReleasesWithFixCallCount()).To(Equal(0))
			Expect(releaseUploader.UploadReleasesCallCount()).To(Equal(0))
			Expect(ui.Said).To(ContainElement("Release-Check for 'ReleaseName/1' has been disabled."))
		})

		It("returns error and does not deploy if uploading releases fails", func() {
			deployOpts.Args.Manifest = opts.FileBytesArg{
				Bytes: []byte(`
name: dep
releases:
- name: capi
  sha1: capi-sha1
  url: https://capi-url
  version: 1+capi
`),
			}

			releaseUploader.UploadReleasesReturns(nil, errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))

			Expect(deployment.UpdateCallCount()).To(Equal(0))
		})

		It("uploads releases but does not deploy if confirmation is rejected", func() {
			deployOpts.Args.Manifest = opts.FileBytesArg{
				Bytes: []byte(`
name: dep
releases:
- name: capi
  sha1: capi-sha1
  url: /capi-url
  version: create
`),
			}

			ui.AskedConfirmationErr = errors.New("stop")

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("stop"))

			Expect(releaseUploader.UploadReleasesCallCount()).To(Equal(1))
			Expect(deployment.UpdateCallCount()).To(Equal(0))
		})

		It("returns an error if diffing failed", func() {
			deployment.DiffReturns(boshdir.DeploymentDiff{}, errors.New("Fetching diff result"))

			err := act()
			Expect(err).To(HaveOccurred())
		})

		It("gets the diff from the deployment", func() {
			diff := [][]interface{}{
				[]interface{}{"some line that stayed", nil},
				[]interface{}{"some line that was added", "added"},
				[]interface{}{"some line that was removed", "removed"},
			}

			expectedDiff := boshdir.NewDeploymentDiff(diff, nil)
			deployment.DiffReturns(expectedDiff, nil)
			err := act()
			Expect(err).ToNot(HaveOccurred())
			Expect(deployment.DiffCallCount()).To(Equal(1))
			Expect(ui.Said).To(ContainElement("  some line that stayed\n"))
			Expect(ui.Said).To(ContainElement("+ some line that was added\n"))
			Expect(ui.Said).To(ContainElement("- some line that was removed\n"))
		})

		It("deploys manifest with diff context", func() {
			context := map[string]interface{}{
				"cloud_config_id":   2,
				"runtime_config_id": 3,
			}
			expectedDiff := boshdir.NewDeploymentDiff(nil, context)

			deployment.DiffReturns(expectedDiff, nil)
			err := act()
			Expect(err).ToNot(HaveOccurred())
			Expect(deployment.DiffCallCount()).To(Equal(1))

			_, updateOptions := deployment.UpdateArgsForCall(0)
			Expect(updateOptions.Diff).To(Equal(expectedDiff))
		})

		It("returns error if deploying failed", func() {
			deployment.UpdateReturns(errors.New("fake-err"))

			err := act()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("overwrites the deployOpts with the flags from configs of type deploy", func() {
			configs := []boshdir.Config{
				{
					ID:   "1",
					Name: "default", Type: "deploy",
					CreatedAt: "0000",
					Team:      "",
					Content:   "flags:\n  - fix",
					Current:   true,
				},
			}

			director.ListConfigsReturns(configs, nil)

			err := act()
			Expect(err).ToNot(HaveOccurred())
			Expect(deployment.UpdateCallCount()).To(Equal(1))

			_, updateOpts := deployment.UpdateArgsForCall(0)

			Expect(updateOpts).To(Equal(boshdir.UpdateOpts{
				Fix: true,
			}))
		})

		It("overwrites the deployOpts with the flags from configs of type deploy if the deployment is included", func() {
			configs := []boshdir.Config{
				{
					ID:        "1",
					Name:      "default",
					Type:      "deploy",
					CreatedAt: "0000",
					Team:      "",
					Content:   "flags:\n  - fix\ninclude:\n  - dep",
					Current:   true,
				},
			}

			director.ListConfigsReturns(configs, nil)
			deployment.NameReturns("dep")

			err := act()
			Expect(err).ToNot(HaveOccurred())
			Expect(deployment.UpdateCallCount()).To(Equal(1))

			_, updateOpts := deployment.UpdateArgsForCall(0)

			Expect(updateOpts).To(Equal(boshdir.UpdateOpts{
				Fix: true,
			}))
		})

		It("does not overwrite the deployOpts with the flags from configs of type deploy if the deployment is not included", func() {
			configs := []boshdir.Config{
				{
					ID:        "1",
					Name:      "default",
					Type:      "deploy",
					CreatedAt: "0000",
					Team:      "",
					Content:   "flags:\n  - fix\ninclude:\n  - foo",
					Current:   true,
				},
			}

			director.ListConfigsReturns(configs, nil)
			deployment.NameReturns("dep")

			err := act()
			Expect(err).ToNot(HaveOccurred())
			Expect(deployment.UpdateCallCount()).To(Equal(1))

			_, updateOpts := deployment.UpdateArgsForCall(0)

			Expect(updateOpts).To(Equal(boshdir.UpdateOpts{
				Fix: false,
			}))
		})

		It("does not overwrite the deployOpts with the flags from configs of type deploy if the deployment is excluded", func() {
			configs := []boshdir.Config{
				{
					ID:        "1",
					Name:      "default",
					Type:      "deploy",
					CreatedAt: "0000",
					Team:      "",
					Content:   "flags:\n  - fix\nexclude:\n  - dep",
					Current:   true,
				},
			}

			director.ListConfigsReturns(configs, nil)
			deployment.NameReturns("dep")

			err := act()
			Expect(err).ToNot(HaveOccurred())
			Expect(deployment.UpdateCallCount()).To(Equal(1))

			_, updateOpts := deployment.UpdateArgsForCall(0)

			Expect(updateOpts).To(Equal(boshdir.UpdateOpts{
				Fix: false,
			}))
		})

		It("overwrites the deployOpts with the flags from configs of type deploy if the deployment is not excluded", func() {
			configs := []boshdir.Config{
				{
					ID:        "1",
					Name:      "default",
					Type:      "deploy",
					CreatedAt: "0000",
					Team:      "",
					Content:   "flags:\n  - fix\nexclude:\n  - foo",
					Current:   true,
				},
			}

			director.ListConfigsReturns(configs, nil)
			deployment.NameReturns("dep")

			err := act()
			Expect(err).ToNot(HaveOccurred())
			Expect(deployment.UpdateCallCount()).To(Equal(1))

			_, updateOpts := deployment.UpdateArgsForCall(0)

			Expect(updateOpts).To(Equal(boshdir.UpdateOpts{
				Fix: true,
			}))
		})
	})
})
