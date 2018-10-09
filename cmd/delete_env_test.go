package cmd_test

import (
	bicmd "github.com/cloudfoundry/bosh-cli/cmd"
	"github.com/cppforlife/go-patch/patch"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mock_cmd "github.com/cloudfoundry/bosh-cli/cmd/mocks"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	"github.com/golang/mock/gomock"

	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	fakebiui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
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
			mockDeploymentDeleter *mock_cmd.MockDeploymentDeleter
			fs                    *fakesys.FakeFileSystem

			fakeUI                 *fakeui.FakeUI
			fakeStage              *fakebiui.FakeStage
			deploymentManifestPath = "/deployment-dir/fake-deployment-manifest.yml"
			statePath              string
			skipDrain              bool
		)

		var newDeleteEnvCmd = func() *bicmd.DeleteEnvCmd {
			doGetFunc := func(manifestPath string, statePath_ string, vars boshtpl.Variables, op patch.Op) bicmd.DeploymentDeleter {
				Expect(manifestPath).To(Equal(deploymentManifestPath))
				Expect(vars).To(Equal(boshtpl.NewMultiVars([]boshtpl.Variables{boshtpl.StaticVariables{"key": "value"}})))
				Expect(op).To(Equal(patch.Ops{patch.ErrOp{}}))
				statePath = statePath_
				return mockDeploymentDeleter
			}

			return bicmd.NewDeleteEnvCmd(fakeUI, doGetFunc)
		}

		var writeDeploymentManifest = func() {
			fs.WriteFileString(deploymentManifestPath, `---manifest-content`)
		}

		BeforeEach(func() {
			mockDeploymentDeleter = mock_cmd.NewMockDeploymentDeleter(mockCtrl)
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
				newDeleteEnvCmd().Run(fakeStage, bicmd.DeleteEnvOpts{
					Args: bicmd.DeleteEnvArgs{
						Manifest: bicmd.FileBytesWithPathArg{Path: deploymentManifestPath},
					},
					SkipDrain: skipDrain,
					VarFlags: bicmd.VarFlags{
						VarKVs: []boshtpl.VarKV{{Name: "key", Value: "value"}},
					},
					OpsFlags: bicmd.OpsFlags{
						OpsFiles: []bicmd.OpsFileArg{
							{Ops: patch.Ops([]patch.Op{patch.ErrOp{}})},
						},
					},
				})
			})
		})

		Context("state path is NOT specified", func() {
			It("sends the manifest on to the deleter", func() {
				mockDeploymentDeleter.EXPECT().DeleteDeployment(skipDrain, fakeStage).Return(nil)
				newDeleteEnvCmd().Run(fakeStage, bicmd.DeleteEnvOpts{
					Args: bicmd.DeleteEnvArgs{
						Manifest: bicmd.FileBytesWithPathArg{Path: deploymentManifestPath},
					},
					SkipDrain: skipDrain,
					VarFlags: bicmd.VarFlags{
						VarKVs: []boshtpl.VarKV{{Name: "key", Value: "value"}},
					},
					OpsFlags: bicmd.OpsFlags{
						OpsFiles: []bicmd.OpsFileArg{
							{Ops: patch.Ops([]patch.Op{patch.ErrOp{}})},
						},
					},
				})

				Expect(statePath).To(Equal(""))
			})
		})

		Context("state path is specified", func() {
			It("sends the manifest on to the deleter", func() {
				mockDeploymentDeleter.EXPECT().DeleteDeployment(skipDrain, fakeStage).Return(nil)
				newDeleteEnvCmd().Run(fakeStage, bicmd.DeleteEnvOpts{
					StatePath: "/new/state/file/path/state.json",
					SkipDrain: skipDrain,
					Args: bicmd.DeleteEnvArgs{
						Manifest: bicmd.FileBytesWithPathArg{Path: deploymentManifestPath},
					},
					VarFlags: bicmd.VarFlags{
						VarKVs: []boshtpl.VarKV{{Name: "key", Value: "value"}},
					},
					OpsFlags: bicmd.OpsFlags{
						OpsFiles: []bicmd.OpsFileArg{
							{Ops: patch.Ops([]patch.Op{patch.ErrOp{}})},
						},
					},
				})

				Expect(statePath).To(Equal("/new/state/file/path/state.json"))
			})
		})

		Context("when the deployment deleter returns an error", func() {
			It("sends the manifest on to the deleter", func() {
				err := bosherr.Error("boom")
				mockDeploymentDeleter.EXPECT().DeleteDeployment(skipDrain, fakeStage).Return(err)
				returnedErr := newDeleteEnvCmd().Run(fakeStage, bicmd.DeleteEnvOpts{
					Args: bicmd.DeleteEnvArgs{
						Manifest: bicmd.FileBytesWithPathArg{Path: deploymentManifestPath},
					},
					SkipDrain: skipDrain,
					VarFlags: bicmd.VarFlags{
						VarKVs: []boshtpl.VarKV{{Name: "key", Value: "value"}},
					},
					OpsFlags: bicmd.OpsFlags{
						OpsFiles: []bicmd.OpsFileArg{
							{Ops: patch.Ops([]patch.Op{patch.ErrOp{}})},
						},
					},
				})
				Expect(returnedErr).To(Equal(err))
			})
		})
	})
})
