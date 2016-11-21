package cmd_test

import (
	bicmd "github.com/cloudfoundry/bosh-cli/cmd"
	"github.com/cppforlife/go-patch/patch"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mock_cmd "github.com/cloudfoundry/bosh-cli/cmd/mocks"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	"github.com/golang/mock/gomock"

	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	fakebiui "github.com/cloudfoundry/bosh-cli/ui/fakes"
	fakeui "github.com/cloudfoundry/bosh-cli/ui/fakes"
)

var _ = Describe("DeleteCmd", func() {
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
			logger                boshlog.Logger

			fakeUI                 *fakeui.FakeUI
			fakeStage              *fakebiui.FakeStage
			deploymentManifestPath = "/deployment-dir/fake-deployment-manifest.yml"
			stateFilePath          string
		)

		var newDeleteCmd = func() *bicmd.DeleteCmd {
			doGetFunc := func(manifestPath string, statePath string, vars boshtpl.Variables, op patch.Op) bicmd.DeploymentDeleter {
				Expect(manifestPath).To(Equal(deploymentManifestPath))
				Expect(vars).To(Equal(boshtpl.NewMultiVars([]boshtpl.Variables{boshtpl.StaticVariables{"key": "value"}})))
				Expect(op).To(Equal(patch.Ops{patch.ErrOp{}}))
				stateFilePath = statePath
				return mockDeploymentDeleter
			}

			return bicmd.NewDeleteCmd(fakeUI, doGetFunc)
		}

		var writeDeploymentManifest = func() {
			fs.WriteFileString(deploymentManifestPath, `---manifest-content`)
		}

		BeforeEach(func() {
			mockDeploymentDeleter = mock_cmd.NewMockDeploymentDeleter(mockCtrl)
			fs = fakesys.NewFakeFileSystem()
			fs.EnableStrictTempRootBehavior()
			logger = boshlog.NewLogger(boshlog.LevelNone)
			fakeUI = &fakeui.FakeUI{}
			writeDeploymentManifest()
		})

		Context("state path is NOT specified", func() {
			It("sends the manifest on to the deleter", func() {
				mockDeploymentDeleter.EXPECT().DeleteDeployment(fakeStage).Return(nil)
				newDeleteCmd().Run(fakeStage, bicmd.DeleteEnvOpts{
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

				Expect(stateFilePath).To(Equal(""))
			})
		})

		Context("state path is specified", func() {
			statePath := "/new/state/file/path/state.json"

			It("sends the manifest on to the deleter", func() {
				mockDeploymentDeleter.EXPECT().DeleteDeployment(fakeStage).Return(nil)
				newDeleteCmd().Run(fakeStage, bicmd.DeleteEnvOpts{
					StateFile: &statePath,
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

				Expect(stateFilePath).To(Equal(statePath))
			})
		})

		Context("when the deployment deleter returns an error", func() {
			It("sends the manifest on to the deleter", func() {
				err := bosherr.Error("boom")
				mockDeploymentDeleter.EXPECT().DeleteDeployment(fakeStage).Return(err)
				returnedErr := newDeleteCmd().Run(fakeStage, bicmd.DeleteEnvOpts{
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
				Expect(returnedErr).To(Equal(err))
			})
		})
	})
})
