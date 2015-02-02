package pkg_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/installation/pkg"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"

	bmproperty "github.com/cloudfoundry/bosh-micro-cli/common/property"
	bminstallmanifest "github.com/cloudfoundry/bosh-micro-cli/installation/manifest"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakebmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"
	fakebmcomp "github.com/cloudfoundry/bosh-micro-cli/installation/pkg/fakes"
	fakebmtemp "github.com/cloudfoundry/bosh-micro-cli/templatescompiler/fakes"
)

var _ = Describe("ReleaseCompiler", func() {
	var (
		fakeReleasePackagesCompiler *fakebmcomp.FakeReleasePackagesCompiler
		fakeTemplatesCompiler       *fakebmtemp.FakeTemplatesCompiler
		fakeFS                      *fakesys.FakeFileSystem
		releaseCompiler             ReleaseCompiler
		cpiJob                      bmrel.Job
		release                     bmrel.Release
		logger                      boshlog.Logger
	)

	BeforeEach(func() {
		fakeReleasePackagesCompiler = fakebmcomp.NewFakeReleasePackagesCompiler()
		fakeTemplatesCompiler = fakebmtemp.NewFakeTemplatesCompiler()

		logger = boshlog.NewLogger(boshlog.LevelNone)

		releaseCompiler = NewReleaseCompiler(
			fakeReleasePackagesCompiler,
			fakeTemplatesCompiler,
			logger,
		)

		fakeFS = fakesys.NewFakeFileSystem()
		cpiJob = bmrel.Job{
			Name: "cpi",
		}
		release = bmrel.NewRelease(
			"fake-release-name",
			"fake-version",
			[]bmrel.Job{cpiJob},
			[]*bmrel.Package{},
			"/some/release/path",
			fakeFS,
		)
	})

	Describe("Compile", func() {
		var (
			manifest            bminstallmanifest.Manifest
			deploymentProperies bmproperty.Map
			fakeStage           *fakebmeventlog.FakeStage
		)

		BeforeEach(func() {
			deploymentProperies = bmproperty.Map{
				"fake-property-key": "fake-property-value",
			}

			manifest = bminstallmanifest.Manifest{
				Name: "fake-deployment-name",
				RawProperties: map[interface{}]interface{}{
					"fake-property-key": "fake-property-value",
				},
			}

			fakeStage = fakebmeventlog.NewFakeStage()

			fakeTemplatesCompiler.SetCompileBehavior([]bmrel.Job{cpiJob}, "fake-deployment-name", deploymentProperies, fakeStage, nil)
		})

		It("compiles the release", func() {
			err := releaseCompiler.Compile(release, manifest, fakeStage)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeReleasePackagesCompiler.CompileRelease.Name()).To(Equal("fake-release-name"))
			Expect(fakeReleasePackagesCompiler.CompileStage).To(Equal(fakeStage))
		})

		It("compiles templates", func() {
			err := releaseCompiler.Compile(release, manifest, fakeStage)
			Expect(err).NotTo(HaveOccurred())
			Expect(fakeTemplatesCompiler.CompileInputs).To(HaveLen(1))
			Expect(fakeTemplatesCompiler.CompileInputs[0]).To(Equal(fakebmtemp.CompileInput{
				Jobs:                 release.Jobs(),
				DeploymentName:       "fake-deployment-name",
				DeploymentProperties: deploymentProperies,
				Stage:                fakeStage,
			}))
		})

		Context("when packages compilation fails", func() {
			It("returns error", func() {
				fakeReleasePackagesCompiler.CompileError = errors.New("fake-compile-error")
				err := releaseCompiler.Compile(release, manifest, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-compile-error"))
			})
		})

		Context("when cpi release has no job named 'cpi'", func() {
			It("returns error", func() {
				release.Jobs()[0].Name = "not-the-cpi"
				err := releaseCompiler.Compile(release, manifest, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Job 'cpi' not found in release 'fake-release-name'"))
			})
		})

		Context("when compiling templates fails", func() {
			BeforeEach(func() {
				err := errors.New("fake-compiling-templates-error")
				fakeTemplatesCompiler.SetCompileBehavior(release.Jobs(), "fake-deployment-name", deploymentProperies, fakeStage, err)
			})

			It("returns an error", func() {
				err := releaseCompiler.Compile(release, manifest, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-compiling-templates-error"))
			})
		})

		Context("when parsing properties fails", func() {
			BeforeEach(func() {
				manifest.RawProperties = map[interface{}]interface{}{
					123: "fake-property-value",
				}
			})

			It("returns an error", func() {
				err := releaseCompiler.Compile(release, manifest, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Getting installation manifest properties"))
			})
		})
	})
})
