package deployer_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-micro-cli/deployer"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	bmdepl "github.com/cloudfoundry/bosh-micro-cli/deployment"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakebmcomp "github.com/cloudfoundry/bosh-micro-cli/compile/fakes"
	fakebmrel "github.com/cloudfoundry/bosh-micro-cli/release/fakes"
	fakebmstemcell "github.com/cloudfoundry/bosh-micro-cli/stemcell/fakes"
	testfakes "github.com/cloudfoundry/bosh-micro-cli/testutils/fakes"
	fakebmui "github.com/cloudfoundry/bosh-micro-cli/ui/fakes"
)

var _ = Describe("CpiDeployer", func() {

	var (
		fakeFs               *fakesys.FakeFileSystem
		fakeExtractor        *testfakes.FakeMultiResponseExtractor
		fakeReleaseValidator *fakebmrel.FakeValidator
		fakeReleaseCompiler  *fakebmcomp.FakeReleaseCompiler
		fakeRepo             *fakebmstemcell.FakeRepo
		fakeUI               *fakebmui.FakeUI

		deploymentManifestPath string
		cpiDeployer            CpiDeployer
	)
	BeforeEach(func() {
		fakeFs = fakesys.NewFakeFileSystem()
		fakeExtractor = testfakes.NewFakeMultiResponseExtractor()
		fakeReleaseValidator = fakebmrel.NewFakeValidator()
		fakeReleaseCompiler = fakebmcomp.NewFakeReleaseCompiler()
		fakeRepo = fakebmstemcell.NewFakeRepo()
		fakeUI = &fakebmui.FakeUI{}
		logger := boshlog.NewLogger(boshlog.LevelNone)

		deploymentManifestPath = "/fake/manifest.yml"
		cpiDeployer = NewCpiDeployer(fakeUI, fakeFs, fakeExtractor, fakeReleaseValidator, fakeReleaseCompiler, logger)
	})

	Describe("Deploy", func() {
		var (
			releaseTarballPath string
			deployment         bmdepl.LocalDeployment
		)
		BeforeEach(func() {
			fakeFs.WriteFileString(deploymentManifestPath, "")

			releaseTarballPath = "/fake/release.tgz"

			fakeFs.WriteFileString(releaseTarballPath, "")
			deployment = bmdepl.LocalDeployment{}
		})

		Context("when a extracted release directory can be created", func() {
			var (
				release bmrel.Release
			)

			BeforeEach(func() {
				fakeFs.TempDirDir = "/some/release/path"

				release = bmrel.Release{
					Name:          "fake-release",
					Version:       "fake-version",
					ExtractedPath: "/some/release/path",
					TarballPath:   releaseTarballPath,
				}

				releaseContents :=
					`---
name: fake-release
version: fake-version
`
				fakeFs.WriteFileString("/some/release/path/release.MF", releaseContents)
			})

			Context("and the tarball is a valid BOSH release", func() {
				BeforeEach(func() {
					fakeExtractor.SetDecompressBehavior(releaseTarballPath, "/some/release/path", nil)
					deployment = bmdepl.NewLocalDeployment("fake-deployment-name", map[string]interface{}{})
					fakeReleaseCompiler.SetCompileBehavior(release, deployment, nil)
				})

				It("does not return an error", func() {
					_, err := cpiDeployer.Deploy(deployment, releaseTarballPath)
					Expect(err).NotTo(HaveOccurred())
				})

				It("compiles the release", func() {
					_, err := cpiDeployer.Deploy(deployment, releaseTarballPath)
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeReleaseCompiler.CompileInputs[0].Deployment).To(Equal(deployment))
				})

				It("cleans up the extracted release directory", func() {
					_, err := cpiDeployer.Deploy(deployment, releaseTarballPath)
					Expect(err).NotTo(HaveOccurred())
					Expect(fakeFs.FileExists("/some/release/path")).To(BeFalse())
				})
			})

			Context("and the tarball is not a valid BOSH release", func() {
				BeforeEach(func() {
					fakeExtractor.SetDecompressBehavior(releaseTarballPath, "/some/release/path", nil)
					fakeFs.WriteFileString("/some/release/path/release.MF", `{}`)
					fakeReleaseValidator.ValidateError = errors.New("fake-error")
				})

				It("returns an error", func() {
					_, err := cpiDeployer.Deploy(deployment, releaseTarballPath)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-error"))
				})

				It("cleans up the extracted release directory", func() {
					_, err := cpiDeployer.Deploy(deployment, releaseTarballPath)
					Expect(err).To(HaveOccurred())
					Expect(fakeFs.FileExists("/some/release/path")).To(BeFalse())
				})
			})

			Context("and the tarball cannot be read", func() {
				It("returns an error", func() {
					_, err := cpiDeployer.Deploy(deployment, releaseTarballPath)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("Reading CPI release from `/fake/release.tgz'"))
					Expect(fakeUI.Errors).To(ContainElement("CPI release `/fake/release.tgz' is not a BOSH release"))
				})
			})

			Context("when compilation fails", func() {
				It("returns an error", func() {
					fakeExtractor.SetDecompressBehavior(releaseTarballPath, "/some/release/path", nil)
					fakeReleaseCompiler.SetCompileBehavior(release, deployment, errors.New("fake-compile-error"))

					_, err := cpiDeployer.Deploy(deployment, releaseTarballPath)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake-compile-error"))
					Expect(fakeUI.Errors).To(ContainElement("Could not compile CPI release"))
				})
			})
		})

		Context("when a extracted release path cannot be created", func() {
			BeforeEach(func() {
				fakeFs.TempDirError = errors.New("fake-tmp-dir-error")
			})

			It("returns an error", func() {
				_, err := cpiDeployer.Deploy(deployment, releaseTarballPath)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-tmp-dir-error"))
				Expect(err.Error()).To(ContainSubstring("Creating temp directory"))
				Expect(fakeUI.Errors).To(ContainElement("Could not create a temporary directory"))
			})
		})
	})
})
