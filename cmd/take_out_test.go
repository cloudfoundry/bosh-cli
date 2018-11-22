package cmd_test

import (
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
	fakeout "github.com/cloudfoundry/bosh-cli/takeout/takeoutfakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
)

// bosh take_out cf-6-offline-sources.yml manifest.yml -o use-compiled-release.yml
var _ = Describe("TakeOutCmd", func() {
	var (
		ui      *fakeui.FakeUI
		command TakeOutCmd
		fs      *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}

		fs = fakesys.NewFakeFileSystem()

		command = NewTakeOutCmd(ui, fakeout.FakeUtensils{})
	})

	Describe("Run", func() {
		var (
			opts TakeOutOpts
		)

		BeforeEach(func() {
			file, err := fs.TempFile("take-out")
			if err != nil {
				panic(err)
			}

			opts = TakeOutOpts{
				Args: TakeOutArgs{
					Name:     file.Name(),
					Manifest: FileBytesArg{Bytes: []byte("name: dep")},
				},
			}

		})

		act := func() error { return command.Run(opts) }

		It("manifest", func() {
			err := act()
			Expect(err).ToNot(HaveOccurred())

			//Expect(deployment.UpdateCallCount()).To(Equal(1))

			//bytes, updateOpts := deployment.UpdateArgsForCall(0)
			//Expect(bytes).To(Equal([]byte("name: dep\n")))
			//Expect(updateOpts).To(Equal(boshdir.UpdateOpts{}))
		})
		//
		//		It("deploys manifest allowing to recreate, recreate persistent disks, fix, and skip drain", func() {
		//			opts.RecreatePersistentDisks = true
		//			opts.Recreate = true
		//			opts.Fix = true
		//			opts.SkipDrain = boshdir.SkipDrains{boshdir.SkipDrain{All: true}}
		//
		//			err := act()
		//			Expect(err).ToNot(HaveOccurred())
		//
		//			Expect(deployment.UpdateCallCount()).To(Equal(1))
		//
		//			bytes, updateOpts := deployment.UpdateArgsForCall(0)
		//			Expect(bytes).To(Equal([]byte("name: dep\n")))
		//			Expect(updateOpts).To(Equal(boshdir.UpdateOpts{
		//				RecreatePersistentDisks: true,
		//				Recreate:                true,
		//				Fix:                     true,
		//				SkipDrain:               boshdir.SkipDrains{boshdir.SkipDrain{All: true}},
		//			}))
		//		})
		//
		//		It("deploys manifest allowing to dry_run", func() {
		//			opts.DryRun = true
		//
		//			err := act()
		//			Expect(err).ToNot(HaveOccurred())
		//
		//			Expect(deployment.UpdateCallCount()).To(Equal(1))
		//
		//			bytes, updateOpts := deployment.UpdateArgsForCall(0)
		//			Expect(bytes).To(Equal([]byte("name: dep\n")))
		//			Expect(updateOpts).To(Equal(boshdir.UpdateOpts{
		//				DryRun: true,
		//			}))
		//		})
		//
		//		It("deploys templated manifest", func() {
		//			opts.Args.Manifest = FileBytesArg{
		//				Bytes: []byte("name: dep\nname1: ((name1))\nname2: ((name2))\n"),
		//			}
		//
		//			opts.VarKVs = []boshtpl.VarKV{
		//				{Name: "name1", Value: "val1-from-kv"},
		//			}
		//
		//			opts.VarsFiles = []boshtpl.VarsFileArg{
		//				{Vars: boshtpl.StaticVariables(map[string]interface{}{"name1": "val1-from-file"})},
		//				{Vars: boshtpl.StaticVariables(map[string]interface{}{"name2": "val2-from-file"})},
		//			}
		//
		//			opts.OpsFiles = []OpsFileArg{
		//				{
		//					Ops: patch.Ops([]patch.Op{
		//						patch.ReplaceOp{Path: patch.MustNewPointerFromString("/xyz?"), Value: "val"},
		//					}),
		//				},
		//			}
		//
		//			err := act()
		//			Expect(err).ToNot(HaveOccurred())
		//
		//			Expect(deployment.UpdateCallCount()).To(Equal(1))
		//
		//			bytes, _ := deployment.UpdateArgsForCall(0)
		//			Expect(bytes).To(Equal([]byte("name: dep\nname1: val1-from-kv\nname2: val2-from-file\nxyz: val\n")))
		//		})
		//
		//		It("does not deploy if name specified in the manifest does not match deployment's name", func() {
		//			opts.Args.Manifest = FileBytesArg{
		//				Bytes: []byte("name: other-name"),
		//			}
		//
		//			err := act()
		//			Expect(err).To(HaveOccurred())
		//			Expect(err.Error()).To(Equal(
		//				"Expected manifest to specify deployment name 'dep' but was 'other-name'"))
		//
		//			Expect(deployment.UpdateCallCount()).To(Equal(0))
		//		})
		//
		//		It("uploads releases provided in the manifest after manifest has been interpolated", func() {
		//			opts.Args.Manifest = FileBytesArg{
		//				Bytes: []byte("name: dep\nbefore-upload-manifest: ((key))"),
		//			}
		//
		//			opts.VarKVs = []boshtpl.VarKV{
		//				{Name: "key", Value: "key-val"},
		//			}
		//
		//			releaseUploader.UploadReleasesReturns([]byte("after-upload-manifest"), nil)
		//
		//			err := act()
		//			Expect(err).ToNot(HaveOccurred())
		//
		//			bytes := releaseUploader.UploadReleasesArgsForCall(0)
		//			Expect(bytes).To(Equal([]byte("before-upload-manifest: key-val\nname: dep\n")))
		//
		//			Expect(deployment.UpdateCallCount()).To(Equal(1))
		//
		//			bytes, _ = deployment.UpdateArgsForCall(0)
		//			Expect(bytes).To(Equal([]byte("after-upload-manifest")))
		//		})
		//
		//		It("returns error and does not deploy if uploading releases fails", func() {
		//			opts.Args.Manifest = FileBytesArg{
		//				Bytes: []byte(`
		//name: dep
		//releases:
		//- name: capi
		//  sha1: capi-sha1
		//  url: https://capi-url
		//  version: 1+capi
		//`),
		//			}
		//
		//			releaseUploader.UploadReleasesReturns(nil, errors.New("fake-err"))
		//
		//			err := act()
		//			Expect(err).To(HaveOccurred())
		//			Expect(err.Error()).To(ContainSubstring("fake-err"))
		//
		//			Expect(deployment.UpdateCallCount()).To(Equal(0))
		//		})
		//
		//		It("uploads releases but does not deploy if confirmation is rejected", func() {
		//			opts.Args.Manifest = FileBytesArg{
		//				Bytes: []byte(`
		//name: dep
		//releases:
		//- name: capi
		//  sha1: capi-sha1
		//  url: /capi-url
		//  version: create
		//`),
		//			}
		//
		//			ui.AskedConfirmationErr = errors.New("stop")
		//
		//			err := act()
		//			Expect(err).To(HaveOccurred())
		//			Expect(err.Error()).To(ContainSubstring("stop"))
		//
		//			Expect(releaseUploader.UploadReleasesCallCount()).To(Equal(1))
		//			Expect(deployment.UpdateCallCount()).To(Equal(0))
		//		})
		//
		//		It("returns an error if diffing failed", func() {
		//			deployment.DiffReturns(boshdir.DeploymentDiff{}, errors.New("Fetching diff result"))
		//
		//			err := act()
		//			Expect(err).To(HaveOccurred())
		//		})
		//
		//		It("gets the diff from the deployment", func() {
		//			diff := [][]interface{}{
		//				[]interface{}{"some line that stayed", nil},
		//				[]interface{}{"some line that was added", "added"},
		//				[]interface{}{"some line that was removed", "removed"},
		//			}
		//
		//			expectedDiff := boshdir.NewDeploymentDiff(diff, nil)
		//			deployment.DiffReturns(expectedDiff, nil)
		//			err := act()
		//			Expect(err).ToNot(HaveOccurred())
		//			Expect(deployment.DiffCallCount()).To(Equal(1))
		//			Expect(ui.Said).To(ContainElement("  some line that stayed\n"))
		//			Expect(ui.Said).To(ContainElement("+ some line that was added\n"))
		//			Expect(ui.Said).To(ContainElement("- some line that was removed\n"))
		//		})
		//
		//		It("deploys manifest with diff context", func() {
		//			context := map[string]interface{}{
		//				"cloud_config_id":   2,
		//				"runtime_config_id": 3,
		//			}
		//			expectedDiff := boshdir.NewDeploymentDiff(nil, context)
		//
		//			deployment.DiffReturns(expectedDiff, nil)
		//			err := act()
		//			Expect(err).ToNot(HaveOccurred())
		//			Expect(deployment.DiffCallCount()).To(Equal(1))
		//
		//			_, updateOptions := deployment.UpdateArgsForCall(0)
		//			Expect(updateOptions.Diff).To(Equal(expectedDiff))
		//		})
		//
		//		It("returns error if deploying failed", func() {
		//			deployment.UpdateReturns(errors.New("fake-err"))
		//
		//			err := act()
		//			Expect(err).To(HaveOccurred())
		//			Expect(err.Error()).To(ContainSubstring("fake-err"))
		//		})
	})
})
