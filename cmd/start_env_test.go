package cmd_test

import (
	bicmd "github.com/cloudfoundry/bosh-cli/cmd"
	. "github.com/cloudfoundry/bosh-cli/cmd/opts"
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

var _ = Describe("StartEnvCmd", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Run", func() {
		var (
			mockDeploymentStateManager *mock_cmd.MockDeploymentStateManager
			fs                         *fakesys.FakeFileSystem

			fakeUI                 *fakeui.FakeUI
			fakeStage              *fakebiui.FakeStage
			deploymentManifestPath = "/deployment-dir/fake-deployment-manifest.yml"
			statePath              string
		)

		var newStartEnvCmd = func() *bicmd.StartEnvCmd {
			doGetFunc := func(manifestPath string, statePath_ string, vars boshtpl.Variables, op patch.Op) bicmd.DeploymentStateManager {
				Expect(manifestPath).To(Equal(deploymentManifestPath))
				Expect(vars).To(Equal(boshtpl.NewMultiVars([]boshtpl.Variables{boshtpl.StaticVariables{"key": "value"}})))
				Expect(op).To(Equal(patch.Ops{patch.ErrOp{}}))
				statePath = statePath_
				return mockDeploymentStateManager
			}

			return bicmd.NewStartEnvCmd(fakeUI, doGetFunc)
		}

		var writeDeploymentManifest = func() {
			fs.WriteFileString(deploymentManifestPath, `---manifest-content`)
		}

		BeforeEach(func() {
			mockDeploymentStateManager = mock_cmd.NewMockDeploymentStateManager(mockCtrl)
			fs = fakesys.NewFakeFileSystem()
			fs.EnableStrictTempRootBehavior()
			fakeUI = &fakeui.FakeUI{}
			writeDeploymentManifest()
		})

		Context("state path is NOT specified", func() {
			It("sends the manifest on to the StartDeployment", func() {
				mockDeploymentStateManager.EXPECT().StartDeployment(fakeStage).Return(nil)
				newStartEnvCmd().Run(fakeStage, StartEnvOpts{
					Args: StartStopEnvArgs{
						Manifest: FileBytesWithPathArg{Path: deploymentManifestPath},
					},
					VarFlags: VarFlags{
						VarKVs: []boshtpl.VarKV{{Name: "key", Value: "value"}},
					},
					OpsFlags: OpsFlags{
						OpsFiles: []OpsFileArg{
							{Ops: patch.Ops([]patch.Op{patch.ErrOp{}})},
						},
					},
				})

				Expect(statePath).To(Equal(""))
			})
		})

		Context("state path is specified", func() {
			It("sends the manifest on to the StartDeployment", func() {
				mockDeploymentStateManager.EXPECT().StartDeployment(fakeStage).Return(nil)
				newStartEnvCmd().Run(fakeStage, StartEnvOpts{
					StatePath: "/new/state/file/path/state.json",
					Args: StartStopEnvArgs{
						Manifest: FileBytesWithPathArg{Path: deploymentManifestPath},
					},
					VarFlags: VarFlags{
						VarKVs: []boshtpl.VarKV{{Name: "key", Value: "value"}},
					},
					OpsFlags: OpsFlags{
						OpsFiles: []OpsFileArg{
							{Ops: patch.Ops([]patch.Op{patch.ErrOp{}})},
						},
					},
				})

				Expect(statePath).To(Equal("/new/state/file/path/state.json"))
			})
		})

		Context("when the deployment state changer returns an error", func() {
			It("sends the manifest on to the StartDeployment", func() {
				err := bosherr.Error("boom")
				mockDeploymentStateManager.EXPECT().StartDeployment(fakeStage).Return(err)
				returnedErr := newStartEnvCmd().Run(fakeStage, StartEnvOpts{
					Args: StartStopEnvArgs{
						Manifest: FileBytesWithPathArg{Path: deploymentManifestPath},
					},
					VarFlags: VarFlags{
						VarKVs: []boshtpl.VarKV{{Name: "key", Value: "value"}},
					},
					OpsFlags: OpsFlags{
						OpsFiles: []OpsFileArg{
							{Ops: patch.Ops([]patch.Op{patch.ErrOp{}})},
						},
					},
				})
				Expect(returnedErr).To(Equal(err))
			})
		})
	})
})
