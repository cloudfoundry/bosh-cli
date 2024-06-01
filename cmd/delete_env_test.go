package cmd_test

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	"github.com/cppforlife/go-patch/patch"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	mockcmd "github.com/cloudfoundry/bosh-cli/v7/cmd/mocks"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("DeleteEnvCmd", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Run", func() {
		var (
			mockDeploymentDeleter *mockcmd.MockDeploymentDeleter
			fs                    *fakesys.FakeFileSystem

			fakeUI                 *fakeui.FakeUI
			fakeStage              *fakeui.FakeStage
			deploymentManifestPath = "/deployment-dir/fake-deployment-manifest.yml"
			statePath              string
			skipDrain              bool
		)

		var newDeleteEnvCmd = func() *cmd.DeleteEnvCmd {
			doGetFunc := func(manifestPath string, statePath_ string, vars boshtpl.Variables, op patch.Op) cmd.DeploymentDeleter {
				Expect(manifestPath).To(Equal(deploymentManifestPath))
				Expect(vars).To(Equal(boshtpl.NewMultiVars([]boshtpl.Variables{boshtpl.StaticVariables{"key": "value"}})))
				Expect(op).To(Equal(patch.Ops{patch.ErrOp{}}))
				statePath = statePath_
				return mockDeploymentDeleter
			}

			return cmd.NewDeleteEnvCmd(fakeUI, doGetFunc)
		}

		var writeDeploymentManifest = func() {
			err := fs.WriteFileString(deploymentManifestPath, `---manifest-content`)
			Expect(err).ToNot(HaveOccurred())
		}

		BeforeEach(func() {
			mockDeploymentDeleter = mockcmd.NewMockDeploymentDeleter(mockCtrl)
			fs = fakesys.NewFakeFileSystem()
			fs.EnableStrictTempRootBehavior()
			fakeUI = &fakeui.FakeUI{}
			writeDeploymentManifest()
			skipDrain = false
		})

		Context("when skip drain is specified", func() {
			It("gets passed to DeleteDeployment", func() {
				skipDrain = true
				mockDeploymentDeleter.EXPECT().DeleteDeployment(skipDrain, fakeStage).Return(nil)
				err := newDeleteEnvCmd().Run(fakeStage, opts.DeleteEnvOpts{
					Args: opts.DeleteEnvArgs{
						Manifest: opts.FileBytesWithPathArg{Path: deploymentManifestPath},
					},
					SkipDrain: skipDrain,
					VarFlags: opts.VarFlags{
						VarKVs: []boshtpl.VarKV{{Name: "key", Value: "value"}},
					},
					OpsFlags: opts.OpsFlags{
						OpsFiles: []opts.OpsFileArg{
							{Ops: []patch.Op{patch.ErrOp{}}},
						},
					},
				})
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("state path is NOT specified", func() {
			It("sends the manifest on to the deleter", func() {
				mockDeploymentDeleter.EXPECT().DeleteDeployment(skipDrain, fakeStage).Return(nil)
				err := newDeleteEnvCmd().Run(fakeStage, opts.DeleteEnvOpts{
					Args: opts.DeleteEnvArgs{
						Manifest: opts.FileBytesWithPathArg{Path: deploymentManifestPath},
					},
					SkipDrain: skipDrain,
					VarFlags: opts.VarFlags{
						VarKVs: []boshtpl.VarKV{{Name: "key", Value: "value"}},
					},
					OpsFlags: opts.OpsFlags{
						OpsFiles: []opts.OpsFileArg{
							{Ops: []patch.Op{patch.ErrOp{}}},
						},
					},
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(statePath).To(Equal(""))
			})
		})

		Context("state path is specified", func() {
			It("sends the manifest on to the deleter", func() {
				mockDeploymentDeleter.EXPECT().DeleteDeployment(skipDrain, fakeStage).Return(nil)
				err := newDeleteEnvCmd().Run(fakeStage, opts.DeleteEnvOpts{
					StatePath: "/new/state/file/path/state.json",
					SkipDrain: skipDrain,
					Args: opts.DeleteEnvArgs{
						Manifest: opts.FileBytesWithPathArg{Path: deploymentManifestPath},
					},
					VarFlags: opts.VarFlags{
						VarKVs: []boshtpl.VarKV{{Name: "key", Value: "value"}},
					},
					OpsFlags: opts.OpsFlags{
						OpsFiles: []opts.OpsFileArg{
							{Ops: []patch.Op{patch.ErrOp{}}},
						},
					},
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(statePath).To(Equal("/new/state/file/path/state.json"))
			})
		})

		Context("when the deployment deleter returns an error", func() {
			It("sends the manifest on to the deleter", func() {
				err := bosherr.Error("boom")
				mockDeploymentDeleter.EXPECT().DeleteDeployment(skipDrain, fakeStage).Return(err)
				returnedErr := newDeleteEnvCmd().Run(fakeStage, opts.DeleteEnvOpts{
					Args: opts.DeleteEnvArgs{
						Manifest: opts.FileBytesWithPathArg{Path: deploymentManifestPath},
					},
					SkipDrain: skipDrain,
					VarFlags: opts.VarFlags{
						VarKVs: []boshtpl.VarKV{{Name: "key", Value: "value"}},
					},
					OpsFlags: opts.OpsFlags{
						OpsFiles: []opts.OpsFileArg{
							{Ops: []patch.Op{patch.ErrOp{}}},
						},
					},
				})
				Expect(returnedErr).To(Equal(err))
			})
		})
	})
})
