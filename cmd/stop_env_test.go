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

var _ = Describe("StopEnvCmd", func() {
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Run", func() {
		var (
			mockDeploymentStateManager *mockcmd.MockDeploymentStateManager
			fs                         *fakesys.FakeFileSystem

			fakeUI                 *fakeui.FakeUI
			fakeStage              *fakeui.FakeStage
			deploymentManifestPath = "/deployment-dir/fake-deployment-manifest.yml"
			statePath              string
			skipDrain              bool
		)

		var newStopEnvCmd = func() *cmd.StopEnvCmd {
			doGetFunc := func(manifestPath string, statePath_ string, vars boshtpl.Variables, op patch.Op) cmd.DeploymentStateManager {
				Expect(manifestPath).To(Equal(deploymentManifestPath))
				Expect(vars).To(Equal(boshtpl.NewMultiVars([]boshtpl.Variables{boshtpl.StaticVariables{"key": "value"}})))
				Expect(op).To(Equal(patch.Ops{patch.ErrOp{}}))
				statePath = statePath_
				return mockDeploymentStateManager
			}

			return cmd.NewStopEnvCmd(fakeUI, doGetFunc)
		}

		var writeDeploymentManifest = func() {
			err := fs.WriteFileString(deploymentManifestPath, `---manifest-content`)
			Expect(err).ToNot(HaveOccurred())
		}

		BeforeEach(func() {
			mockDeploymentStateManager = mockcmd.NewMockDeploymentStateManager(mockCtrl)
			fs = fakesys.NewFakeFileSystem()
			fs.EnableStrictTempRootBehavior()
			fakeUI = &fakeui.FakeUI{}
			writeDeploymentManifest()
			skipDrain = false
		})

		Context("when skip drain is specified", func() {
			It("gets passed to StopDeployment", func() {
				skipDrain = true
				mockDeploymentStateManager.EXPECT().StopDeployment(skipDrain, fakeStage).Return(nil)
				err := newStopEnvCmd().Run(fakeStage, opts.StopEnvOpts{
					Args: opts.StartStopEnvArgs{
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
			It("sends the manifest on to the StopDeployment", func() {
				mockDeploymentStateManager.EXPECT().StopDeployment(skipDrain, fakeStage).Return(nil)
				err := newStopEnvCmd().Run(fakeStage, opts.StopEnvOpts{
					Args: opts.StartStopEnvArgs{
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
			It("sends the manifest on to the StopDeployment", func() {
				mockDeploymentStateManager.EXPECT().StopDeployment(skipDrain, fakeStage).Return(nil)
				err := newStopEnvCmd().Run(fakeStage, opts.StopEnvOpts{
					StatePath: "/new/state/file/path/state.json",
					SkipDrain: skipDrain,
					Args: opts.StartStopEnvArgs{
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

		Context("when the deployment state changer returns an error", func() {
			It("sends the manifest on to the StopDeployment", func() {
				err := bosherr.Error("boom")
				mockDeploymentStateManager.EXPECT().StopDeployment(skipDrain, fakeStage).Return(err)
				returnedErr := newStopEnvCmd().Run(fakeStage, opts.StopEnvOpts{
					Args: opts.StartStopEnvArgs{
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
