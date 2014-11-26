package compile_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmcomp "github.com/cloudfoundry/bosh-micro-cli/cpi/compile/fakes"
	fakebmtemp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/fakes"

	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	. "github.com/cloudfoundry/bosh-micro-cli/cpi/compile"
)

var _ = Describe("ReleaseCompiler", func() {
	var (
		fakeReleasePackagesCompiler *fakebmcomp.FakeReleasePackagesCompiler
		fakeTemplatesCompiler       *fakebmtemp.FakeTemplatesCompiler
		fakeFS                      *fakesys.FakeFileSystem
		releaseCompiler             ReleaseCompiler
		release                     bmrel.Release
	)

	BeforeEach(func() {
		fakeReleasePackagesCompiler = fakebmcomp.NewFakeReleasePackagesCompiler()
		fakeTemplatesCompiler = fakebmtemp.NewFakeTemplatesCompiler()

		releaseCompiler = NewReleaseCompiler(
			fakeReleasePackagesCompiler,
			fakeTemplatesCompiler,
		)

		fakeFS = fakesys.NewFakeFileSystem()
		release = bmrel.NewRelease(
			"fake-release-name",
			"fake-version",
			[]bmrel.Job{
				bmrel.Job{
					Name: "fake-job-name",
				},
			},
			[]*bmrel.Package{},
			"/some/release/path",
			fakeFS,
		)
	})

	Describe("Compile", func() {
		var (
			deployment          bmdepl.CPIDeployment
			deploymentProperies map[string]interface{}
		)

		BeforeEach(func() {
			deploymentProperies = map[string]interface{}{
				"fake-property-key": "fake-property-value",
			}

			deployment = bmdepl.CPIDeployment{
				Name: "fake-deployment-name",
				RawProperties: map[interface{}]interface{}{
					"fake-property-key": "fake-property-value",
				},
				Jobs: []bmdepl.Job{},
			}
			fakeTemplatesCompiler.SetCompileBehavior(release.Jobs(), "fake-deployment-name", deploymentProperies, nil)
		})

		It("compiles the release", func() {
			err := releaseCompiler.Compile(release, deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeReleasePackagesCompiler.CompileRelease.Name()).To(Equal("fake-release-name"))
		})

		It("compiles templates", func() {
			err := releaseCompiler.Compile(release, deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeTemplatesCompiler.CompileInputs).To(HaveLen(1))
			Expect(fakeTemplatesCompiler.CompileInputs[0]).To(Equal(fakebmtemp.CompileInput{
				Jobs:                 release.Jobs(),
				DeploymentName:       "fake-deployment-name",
				DeploymentProperties: deploymentProperies,
			}))
		})

		Context("when packages compilation fails", func() {
			It("returns error", func() {
				fakeReleasePackagesCompiler.CompileError = errors.New("fake-compile-error")
				err := releaseCompiler.Compile(release, deployment)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-compile-error"))
			})
		})

		Context("when compiling templates fails", func() {
			BeforeEach(func() {
				err := errors.New("fake-compiling-templates-error")
				fakeTemplatesCompiler.SetCompileBehavior(release.Jobs(), "fake-deployment-name", deploymentProperies, err)
			})

			It("returns an error", func() {
				err := releaseCompiler.Compile(release, deployment)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-compiling-templates-error"))
			})
		})

		Context("when parsing properties fails", func() {
			BeforeEach(func() {
				deployment.RawProperties = map[interface{}]interface{}{
					123: "fake-property-value",
				}
			})

			It("returns an error", func() {
				err := releaseCompiler.Compile(release, deployment)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Getting deployment properties"))
			})
		})
	})
})
