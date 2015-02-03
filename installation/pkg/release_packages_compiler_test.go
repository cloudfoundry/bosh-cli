package pkg_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	faketime "github.com/cloudfoundry/bosh-agent/time/fakes"

	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
	bmreljob "github.com/cloudfoundry/bosh-micro-cli/release/job"
	bmrelpkg "github.com/cloudfoundry/bosh-micro-cli/release/pkg"

	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"
	fakebminstallpkg "github.com/cloudfoundry/bosh-micro-cli/installation/pkg/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/installation/pkg"
)

var _ = Describe("ReleaseCompiler", func() {
	var (
		release                 bmrel.Release
		releasePackagesCompiler ReleasePackagesCompiler
		packageCompiler         *fakebminstallpkg.FakePackageCompiler
		eventLogger             *fakebmlog.FakeEventLogger
		fakeStage               *fakebmlog.FakeStage
		timeService             *faketime.FakeService
		fakeFS                  *fakesys.FakeFileSystem
	)

	BeforeEach(func() {
		packageCompiler = fakebminstallpkg.NewFakePackageCompiler()
		eventLogger = fakebmlog.NewFakeEventLogger()
		fakeStage = fakebmlog.NewFakeStage()
		timeService = &faketime.FakeService{}
		releasePackagesCompiler = NewReleasePackagesCompiler(packageCompiler, eventLogger, timeService)
		fakeFS = fakesys.NewFakeFileSystem()
		release = bmrel.NewRelease(
			"fake-release",
			"fake-version",
			[]bmreljob.Job{},
			[]*bmrelpkg.Package{},
			"/some/release/path",
			fakeFS,
		)
	})

	Context("Compile", func() {
		Context("when there is a release", func() {
			var expectedPackages []*bmrelpkg.Package
			var package1, package2 bmrelpkg.Package

			BeforeEach(func() {
				package1 = bmrelpkg.Package{Name: "fake-package-1", Fingerprint: "fake-fingerprint-1", Dependencies: []*bmrelpkg.Package{}}
				package2 = bmrelpkg.Package{Name: "fake-package-2", Fingerprint: "fake-fingerprint-2", Dependencies: []*bmrelpkg.Package{&package1}}

				expectedPackages = []*bmrelpkg.Package{&package1, &package2}

				release = bmrel.NewRelease(
					"fake-release",
					"fake-version",
					[]bmreljob.Job{},
					[]*bmrelpkg.Package{&package2, &package1},
					"/some/release/path",
					fakeFS,
				)
			})

			It("compiles each package", func() {
				err := releasePackagesCompiler.Compile(release, fakeStage)
				Expect(err).NotTo(HaveOccurred())
				Expect(packageCompiler.CompilePackages).To(Equal(expectedPackages))
			})

			It("compiles each package and returns error for first package", func() {
				packageCompiler.CompileError = bosherr.Error("fake-compilation-error")
				err := releasePackagesCompiler.Compile(release, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-compilation-error"))
			})

			It("logs start and stop events to the eventLogger", func() {
				pkg1Start := time.Now()
				pkg1Finish := pkg1Start.Add(1 * time.Second)
				timeService.NowTimes = []time.Time{pkg1Start, pkg1Finish}

				err := releasePackagesCompiler.Compile(release, fakeStage)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Compiling package 'fake-package-1/fake-fingerprint-1'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Finished,
					},
				}))
			})

			It("logs failure event", func() {
				pkg1Start := time.Now()
				pkg1Fail := pkg1Start.Add(1 * time.Second)
				timeService.NowTimes = []time.Time{pkg1Start, pkg1Fail}

				packageCompiler.CompileError = bosherr.Error("fake-compilation-error")
				err := releasePackagesCompiler.Compile(release, fakeStage)
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "Compiling package 'fake-package-1/fake-fingerprint-1'",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "fake-compilation-error",
				}))
			})

			It("stops compiling after the first failure", func() {
				packageCompiler.CompileError = bosherr.Error("fake-compilation-error")
				err := releasePackagesCompiler.Compile(release, fakeStage)
				Expect(err).To(HaveOccurred())
				Expect(len(packageCompiler.CompilePackages)).To(Equal(1))
			})
		})
	})
})
