package compile_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	bmeventlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	faketime "github.com/cloudfoundry/bosh-agent/time/fakes"

	fakebmcomp "github.com/cloudfoundry/bosh-micro-cli/cpideployer/compile/fakes"
	fakebmlog "github.com/cloudfoundry/bosh-micro-cli/eventlogger/fakes"
	fakebmreal "github.com/cloudfoundry/bosh-micro-cli/release/fakes"

	. "github.com/cloudfoundry/bosh-micro-cli/cpideployer/compile"
)

var _ = Describe("ReleaseCompiler", func() {
	var (
		release                 bmrel.Release
		releasePackagesCompiler ReleasePackagesCompiler
		da                      *fakebmreal.FakeDependencyAnalysis
		packageCompiler         *fakebmcomp.FakePackageCompiler
		eventLogger             *fakebmlog.FakeEventLogger
		fakeStage               *fakebmlog.FakeStage
		timeService             *faketime.FakeService
	)

	BeforeEach(func() {
		da = fakebmreal.NewFakeDependencyAnalysis()
		packageCompiler = fakebmcomp.NewFakePackageCompiler()
		eventLogger = fakebmlog.NewFakeEventLogger()
		fakeStage = fakebmlog.NewFakeStage()
		eventLogger.SetNewStageBehavior(fakeStage)
		timeService = &faketime.FakeService{}
		releasePackagesCompiler = NewReleasePackagesCompiler(da, packageCompiler, eventLogger, timeService)
		release = bmrel.Release{}
	})

	Context("Compile", func() {
		It("adds a new event logger stage", func() {
			err := releasePackagesCompiler.Compile(release)
			Expect(err).ToNot(HaveOccurred())

			Expect(eventLogger.NewStageInputs).To(Equal([]fakebmlog.NewStageInput{
				{
					Name: "compiling packages",
				},
			}))

			Expect(fakeStage.Started).To(BeTrue())
			Expect(fakeStage.Finished).To(BeTrue())
		})

		Context("when there is a release", func() {
			var expectedPackages []*bmrel.Package
			var package1, package2 bmrel.Package

			BeforeEach(func() {
				package1 = bmrel.Package{Name: "fake-package-1", Fingerprint: "fake-fingerprint-1"}
				package2 = bmrel.Package{Name: "fake-package-2", Fingerprint: "fake-fingerprint-2"}

				expectedPackages = []*bmrel.Package{&package1, &package2}

				da.DeterminePackageCompilationOrderResult = []*bmrel.Package{
					&package1,
					&package2,
				}
			})

			It("determines the order to compile packages", func() {
				err := releasePackagesCompiler.Compile(release)
				Expect(err).NotTo(HaveOccurred())
				Expect(da.DeterminePackageCompilationOrderRelease).To(Equal(release))
			})

			It("compiles each package", func() {
				err := releasePackagesCompiler.Compile(release)
				Expect(err).NotTo(HaveOccurred())
				Expect(packageCompiler.CompilePackages).To(Equal(expectedPackages))
			})

			It("compiles each package and returns error for first package", func() {
				packageCompiler.CompileError = errors.New("Compilation failed")
				err := releasePackagesCompiler.Compile(release)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Package `fake-package-1' compilation failed"))
			})

			It("logs start and stop events to the eventLogger", func() {
				pkg1Start := time.Now()
				pkg1Finish := pkg1Start.Add(1 * time.Second)
				timeService.NowTimes = []time.Time{pkg1Start, pkg1Finish}

				err := releasePackagesCompiler.Compile(release)
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "fake-package-1/fake-fingerprint-1",
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

				packageCompiler.CompileError = errors.New("Compilation failed")
				err := releasePackagesCompiler.Compile(release)
				Expect(err).To(HaveOccurred())

				Expect(fakeStage.Steps).To(ContainElement(&fakebmlog.FakeStep{
					Name: "fake-package-1/fake-fingerprint-1",
					States: []bmeventlog.EventState{
						bmeventlog.Started,
						bmeventlog.Failed,
					},
					FailMessage: "Compilation failed",
				}))
			})

			It("stops compiling after the first failure", func() {
				packageCompiler.CompileError = errors.New("Compilation failed")
				err := releasePackagesCompiler.Compile(release)
				Expect(err).To(HaveOccurred())
				Expect(len(packageCompiler.CompilePackages)).To(Equal(1))
			})
		})
	})
})
