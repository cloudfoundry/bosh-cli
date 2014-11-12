package compile_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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

		release = bmrel.Release{
			Name:     "fake-release-name",
			Packages: []*bmrel.Package{},
			Jobs: []bmrel.Job{
				bmrel.Job{
					Name: "fake-job-name",
				},
			},
		}
	})

	Describe("Compile", func() {
		var (
			deployment bmdepl.Deployment
		)
		BeforeEach(func() {
			deployment = bmdepl.Deployment{
				Name:          "fake-deployment-name",
				RawProperties: map[interface{}]interface{}{},
				Jobs:          []bmdepl.Job{},
			}
			fakeTemplatesCompiler.SetCompileBehavior(release.Jobs, deployment, nil)
		})

		It("compiles the release", func() {
			err := releaseCompiler.Compile(release, deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeReleasePackagesCompiler.CompileRelease.Name).To(Equal("fake-release-name"))
		})

		It("compiles templates", func() {
			err := releaseCompiler.Compile(release, deployment)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeTemplatesCompiler.CompileInputs).To(HaveLen(1))
			Expect(fakeTemplatesCompiler.CompileInputs[0]).To(Equal(fakebmtemp.CompileInput{
				Jobs:       release.Jobs,
				Deployment: deployment,
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
				fakeTemplatesCompiler.SetCompileBehavior(release.Jobs, deployment, err)
			})

			It("returns an error", func() {
				err := releaseCompiler.Compile(release, deployment)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-compiling-templates-error"))
			})
		})
	})
})
