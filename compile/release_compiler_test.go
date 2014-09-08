package compile_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fakebmcomp "github.com/cloudfoundry/bosh-micro-cli/compile/fakes"
	fakebmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment/fakes"
	fakebmtemp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/fakes"

	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/jobs"

	. "github.com/cloudfoundry/bosh-micro-cli/compile"
)

var _ = Describe("ReleaseCompiler", func() {
	var (
		fakeReleasePackagesCompiler *fakebmcomp.FakeReleasePackagesCompiler
		fakeManifestParser          *fakebmdepl.FakeManifestParser
		fakeTemplatesCompiler       *fakebmtemp.FakeTemplatesCompiler
		releaseCompiler             ReleaseCompiler
		release                     bmrel.Release
		deployment                  bmdepl.Deployment
	)

	BeforeEach(func() {
		fakeReleasePackagesCompiler = fakebmcomp.NewFakeReleasePackagesCompiler()
		fakeManifestParser = fakebmdepl.NewFakeManifestParser()
		fakeTemplatesCompiler = fakebmtemp.NewFakeTemplatesCompiler()

		releaseCompiler = NewReleaseCompiler(
			fakeReleasePackagesCompiler,
			fakeManifestParser,
			fakeTemplatesCompiler,
		)

		release = bmrel.Release{
			Name:     "fake-release-name",
			Packages: []*bmrel.Package{},
			Jobs: []bmreljob.Job{
				bmreljob.Job{
					Name: "fake-job-name",
				},
			},
		}
	})

	Describe("Compile", func() {
		BeforeEach(func() {
			deployment = bmdepl.NewLocalDeployment("fake-deployment-name", map[string]interface{}{})
			fakeManifestParser.SetParseBehavior("/some/deployment/file", deployment, nil)
			fakeTemplatesCompiler.SetCompileBehavior(release.Jobs, deployment, nil)
		})

		It("compiles the release", func() {
			err := releaseCompiler.Compile(release, "/some/deployment/file")
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeReleasePackagesCompiler.CompileRelease.Name).To(Equal("fake-release-name"))
		})

		It("parses deployment manifest", func() {
			err := releaseCompiler.Compile(release, "/some/deployment/file")
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeManifestParser.ParseInputs[0].DeploymentPath).To(Equal("/some/deployment/file"))
		})

		It("compiles templates", func() {
			err := releaseCompiler.Compile(release, "/some/deployment/file")
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
				err := releaseCompiler.Compile(release, "/some/deployment/file")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-compile-error"))
			})
		})

		Context("when parsing deployment manifest fails", func() {
			BeforeEach(func() {
				parserErr := errors.New("fake-manifest-parser-error")
				fakeManifestParser.SetParseBehavior("/some/deployment/file", bmdepl.LocalDeployment{}, parserErr)
			})

			It("returns an error", func() {
				err := releaseCompiler.Compile(release, "/some/deployment/file")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-manifest-parser-error"))
			})
		})

		Context("when compiling templates fails", func() {
			BeforeEach(func() {
				err := errors.New("fake-compiling-templates-error")
				fakeTemplatesCompiler.SetCompileBehavior(release.Jobs, deployment, err)
			})

			It("returns an error", func() {
				err := releaseCompiler.Compile(release, "/some/deployment/file")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-compiling-templates-error"))
			})
		})
	})
})
