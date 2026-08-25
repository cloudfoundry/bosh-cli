package cmd_test

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	"github.com/cppforlife/go-patch/patch"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/cmdfakes"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	"github.com/cloudfoundry/bosh-cli/v7/ui/testui"
)

var _ = Describe("StartEnvCmd", func() {
	Describe("Run", func() {
		var (
			mockDeploymentStateManager *cmdfakes.FakeDeploymentStateManager
			fs                         *fakesys.FakeFileSystem

			TestUI                 *testui.Ui
			fakeStage              *testui.Stage
			deploymentManifestPath = "/deployment-dir/fake-deployment-manifest.yml"
			statePath              string
		)

		var newStartEnvCmd = func() *cmd.StartEnvCmd {
			doGetFunc := func(manifestPath string, statePath_ string, vars boshtpl.Variables, op patch.Op) cmd.DeploymentStateManager {
				Expect(manifestPath).To(Equal(deploymentManifestPath))
				Expect(vars).To(Equal(boshtpl.NewMultiVars([]boshtpl.Variables{boshtpl.StaticVariables{"key": "value"}})))
				Expect(op).To(Equal(patch.Ops{patch.ErrOp{}}))
				statePath = statePath_
				return mockDeploymentStateManager
			}

			return cmd.NewStartEnvCmd(TestUI, doGetFunc)
		}

		var writeDeploymentManifest = func() {
			err := fs.WriteFileString(deploymentManifestPath, `---manifest-content`)
			Expect(err).ToNot(HaveOccurred())
		}

		BeforeEach(func() {
			mockDeploymentStateManager = &cmdfakes.FakeDeploymentStateManager{}
			fs = fakesys.NewFakeFileSystem()
			fs.EnableStrictTempRootBehavior()
			TestUI = &testui.Ui{}
			writeDeploymentManifest()
		})

		Context("state path is NOT specified", func() {
			It("sends the manifest on to the StartDeployment", func() {
				mockDeploymentStateManager.StartDeploymentReturns(nil)
				err := newStartEnvCmd().Run(fakeStage, opts.StartEnvOpts{
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

				Expect(statePath).To(Equal(""))

				Expect(mockDeploymentStateManager.StartDeploymentCallCount()).To(Equal(1))
				Expect(mockDeploymentStateManager.StartDeploymentArgsForCall(0)).To(Equal(fakeStage))
			})
		})

		Context("state path is specified", func() {
			It("sends the manifest on to the StartDeployment", func() {
				mockDeploymentStateManager.StartDeploymentReturns(nil)
				err := newStartEnvCmd().Run(fakeStage, opts.StartEnvOpts{
					StatePath: "/new/state/file/path/state.json",
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
			It("sends the manifest on to the StartDeployment", func() {
				err := bosherr.Error("boom")
				mockDeploymentStateManager.StartDeploymentReturns(err)
				returnedErr := newStartEnvCmd().Run(fakeStage, opts.StartEnvOpts{
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
				Expect(returnedErr).To(Equal(err))
			})
		})
	})
})
