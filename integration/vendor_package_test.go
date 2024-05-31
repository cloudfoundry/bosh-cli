package integration_test

import (
	"fmt"
	"path/filepath"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"github.com/cppforlife/go-patch/patch"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd"
	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
	boshpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
	fakeui "github.com/cloudfoundry/bosh-cli/v7/ui/fakes"
)

var _ = Describe("vendor-package command", func() {
	var (
		ui         *fakeui.FakeUI
		fs         boshsys.FileSystem
		deps       BasicDeps
		cmdFactory Factory
	)

	BeforeEach(func() {
		ui = &fakeui.FakeUI{}
		logger := boshlog.NewLogger(boshlog.LevelNone)
		confUI := boshui.NewWrappingConfUI(ui, logger)

		fs = boshsys.NewOsFileSystem(logger)
		deps = NewBasicDepsWithFS(confUI, fs, logger)
		cmdFactory = NewFactory(deps)
	})

	opFile := func(path string, op patch.Op) {
		contents, err := fs.ReadFile(path)
		Expect(err).ToNot(HaveOccurred())

		tpl := boshtpl.NewTemplate(contents)

		contents, err = tpl.Evaluate(nil, op, boshtpl.EvaluateOpts{})
		Expect(err).ToNot(HaveOccurred())

		err = fs.WriteFile(path, contents)
		Expect(err).ToNot(HaveOccurred())
	}

	findPkg := func(name string, release boshrel.Release) *boshpkg.Package {
		for _, pkg := range release.Packages() {
			if pkg.Name() == name {
				return pkg
			}
		}
		panic(fmt.Sprintf("Expected to find package '%s'", name))
	}

	It("vendors packages", func() {
		upstreamDir, err := fs.TempDir("bosh-vendor-package-int-test")
		Expect(err).ToNot(HaveOccurred())

		defer fs.RemoveAll(upstreamDir) //nolint:errcheck

		{ // Initialize upstream release
			execCmd(cmdFactory, []string{"init-release", "--git", "--dir", upstreamDir})

			blobstoreConfig := fmt.Sprintf(`
blobstore:
  provider: local
  options:
    blobstore_path: %s
`, filepath.Join(upstreamDir, "blobstore"))

			finalConfigPath := filepath.Join(upstreamDir, "config", "final.yml")

			prevContents, err := fs.ReadFileString(finalConfigPath)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(finalConfigPath, prevContents+blobstoreConfig)
			Expect(err).ToNot(HaveOccurred())

			execCmd(cmdFactory, []string{"generate-package", "pkg1", "--dir", upstreamDir})
		}

		{ // Add a bit of content to upstream release
			err := fs.WriteFileString(filepath.Join(upstreamDir, "src", "in-src"), "in-src")
			Expect(err).ToNot(HaveOccurred())

			pkg1SpecPath := filepath.Join(upstreamDir, "packages", "pkg1", "spec")

			replaceOp := patch.ReplaceOp{
				// eq /files/-
				Path: patch.NewPointer([]patch.Token{
					patch.RootToken{},
					patch.KeyToken{Key: "files"},
					patch.AfterLastIndexToken{},
				}),
				Value: "in-src",
			}

			opFile(pkg1SpecPath, replaceOp)

			execCmd(cmdFactory, []string{"create-release", "--final", "--force", "--dir", upstreamDir})
		}

		targetDir, err := fs.TempDir("bosh-vendor-package-int-test")
		Expect(err).ToNot(HaveOccurred())

		defer fs.RemoveAll(targetDir) //nolint:errcheck

		{ // Initialize target release
			execCmd(cmdFactory, []string{"init-release", "--git", "--dir", targetDir})

			blobstoreConfig := fmt.Sprintf(`
blobstore:
  provider: local
  options:
    blobstore_path: %s
`, filepath.Join(targetDir, "blobstore"))

			finalConfigPath := filepath.Join(targetDir, "config", "final.yml")

			prevContents, err := fs.ReadFileString(finalConfigPath)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(finalConfigPath, prevContents+blobstoreConfig)
			Expect(err).ToNot(HaveOccurred())

			execCmd(cmdFactory, []string{"generate-package", "pkg2", "--dir", targetDir})
		}

		{ // Bring over vendored pkg1
			execCmd(cmdFactory, []string{"vendor-package", "pkg1", upstreamDir, "--dir", targetDir})
		}

		{ // Check contents of a target release
			targetTarball, err := fs.TempFile("bosh-vendor-package-int-test")
			Expect(err).ToNot(HaveOccurred())

			defer fs.RemoveAll(targetTarball.Name()) //nolint:errcheck

			execCmd(cmdFactory, []string{"create-release", "--tarball", targetTarball.Name(), "--force", "--dir", targetDir})

			relProvider := boshrel.NewProvider(deps.CmdRunner, deps.Compressor, deps.DigestCalculator, deps.FS, deps.Logger)
			archiveReader := relProvider.NewExtractingArchiveReader()

			release, err := archiveReader.Read(targetTarball.Name())
			Expect(err).ToNot(HaveOccurred())

			defer release.CleanUp() //nolint:errcheck

			pkg1 := release.Packages()[0]
			Expect(fs.ReadFileString(filepath.Join(pkg1.ExtractedPath(), "in-src"))).To(Equal("in-src"))
		}

		{ // Add new bits to upstream release
			err := fs.WriteFileString(filepath.Join(upstreamDir, "src", "in-src"), "in-src-updated")
			Expect(err).ToNot(HaveOccurred())
		}

		{ // Add package dependency to upstream release
			execCmd(cmdFactory, []string{"generate-package", "dependent-pkg", "--dir", upstreamDir})

			err := fs.WriteFileString(filepath.Join(upstreamDir, "src", "dependent-pkg-file"), "in-dependent-pkg")
			Expect(err).ToNot(HaveOccurred())

			specPath := filepath.Join(upstreamDir, "packages", "dependent-pkg", "spec")

			replaceOp := patch.ReplaceOp{
				// eq /files/-
				Path: patch.NewPointer([]patch.Token{
					patch.RootToken{},
					patch.KeyToken{Key: "files"},
					patch.AfterLastIndexToken{},
				}),
				Value: "dependent-pkg-file",
			}

			opFile(specPath, replaceOp)
		}

		{ // Make pkg1 depend on dependent-package
			pkg1SpecPath := filepath.Join(upstreamDir, "packages", "pkg1", "spec")

			replaceOp := patch.ReplaceOp{
				// eq /dependencies/-
				Path: patch.NewPointer([]patch.Token{
					patch.RootToken{},
					patch.KeyToken{Key: "dependencies"},
					patch.AfterLastIndexToken{},
				}),
				Value: "dependent-pkg",
			}

			opFile(pkg1SpecPath, replaceOp)

			execCmd(cmdFactory, []string{"create-release", "--final", "--force", "--dir", upstreamDir})
		}

		{ // Bring over vendored pkg1
			execCmd(cmdFactory, []string{"vendor-package", "pkg1", upstreamDir, "--dir", targetDir})
		}

		{ // Check contents of a target release with updated package version and dependent package
			targetTarball, err := fs.TempFile("bosh-vendor-package-int-test")
			Expect(err).ToNot(HaveOccurred())

			defer fs.RemoveAll(targetTarball.Name()) //nolint:errcheck

			execCmd(cmdFactory, []string{"create-release", "--tarball", targetTarball.Name(), "--force", "--dir", targetDir})

			relProvider := boshrel.NewProvider(deps.CmdRunner, deps.Compressor, deps.DigestCalculator, deps.FS, deps.Logger)
			archiveReader := relProvider.NewExtractingArchiveReader()

			release, err := archiveReader.Read(targetTarball.Name())
			Expect(err).ToNot(HaveOccurred())

			defer release.CleanUp() //nolint:errcheck

			pkg1 := findPkg("pkg1", release)
			content, err := fs.ReadFileString(filepath.Join(pkg1.ExtractedPath(), "in-src"))
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(Equal("in-src-updated"))

			dependentPkg := findPkg("dependent-pkg", release)
			content, err = fs.ReadFileString(filepath.Join(dependentPkg.ExtractedPath(), "dependent-pkg-file"))
			Expect(err).ToNot(HaveOccurred())
			Expect(content).To(Equal("in-dependent-pkg"))

			Expect(pkg1.Dependencies).To(Equal([]*boshpkg.Package{dependentPkg}))
		}
	})
})
