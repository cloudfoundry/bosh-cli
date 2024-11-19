package integration_test

import (
	"fmt"
	"path/filepath"

	"github.com/cppforlife/go-patch/patch"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	boshtpl "github.com/cloudfoundry/bosh-cli/v7/director/template"
	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
	boshpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
)

func findPkg(name string, release boshrel.Release) *boshpkg.Package {
	for _, pkg := range release.Packages() {
		if pkg.Name() == name {
			return pkg
		}
	}
	panic(fmt.Sprintf("Expected to find package '%s'", name))
}

func opFile(path string, op patch.Op) {
	contents, err := fs.ReadFile(path)
	Expect(err).ToNot(HaveOccurred())

	tpl := boshtpl.NewTemplate(contents)

	contents, err = tpl.Evaluate(nil, op, boshtpl.EvaluateOpts{})
	Expect(err).ToNot(HaveOccurred())

	err = fs.WriteFile(path, contents)
	Expect(err).ToNot(HaveOccurred())
}

var _ = Describe("vendor-package command", func() {
	It("vendors packages", func() {
		upstreamDir, err := fs.TempDir("bosh-vendor-package-int-test")
		Expect(err).ToNot(HaveOccurred())

		defer fs.RemoveAll(upstreamDir) //nolint:errcheck

		By("running `init-release` to create the upstream release", func() {
			createAndExecCommand(cmdFactory, []string{"init-release", "--git", "--dir", upstreamDir})

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

			createAndExecCommand(cmdFactory, []string{"generate-package", "pkg1", "--dir", upstreamDir})

			By("adding some content for testing purposes", func() {
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

				createAndExecCommand(cmdFactory, []string{"create-release", "--final", "--force", "--dir", upstreamDir})
			})
		})

		targetDir, err := fs.TempDir("bosh-vendor-package-int-test")
		Expect(err).ToNot(HaveOccurred())

		defer fs.RemoveAll(targetDir) //nolint:errcheck

		By("running `init-release` to create the target release", func() {
			createAndExecCommand(cmdFactory, []string{"init-release", "--git", "--dir", targetDir})

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

			createAndExecCommand(cmdFactory, []string{"generate-package", "pkg2", "--dir", targetDir})
		})

		By("running `vendor-package` to vendor the upstream release's package `pkg1`", func() {
			createAndExecCommand(cmdFactory, []string{"vendor-package", "pkg1", upstreamDir, "--dir", targetDir})
		})

		By("verifying that the upstream release's package `pkg1` has been vendored", func() {
			targetTarball, err := fs.TempFile("bosh-vendor-package-int-test")
			Expect(err).ToNot(HaveOccurred())

			defer fs.RemoveAll(targetTarball.Name()) //nolint:errcheck

			createAndExecCommand(cmdFactory, []string{"create-release", "--tarball", targetTarball.Name(), "--force", "--dir", targetDir})

			relProvider := boshrel.NewProvider(deps.CmdRunner, deps.Compressor, deps.DigestCalculator, deps.FS, deps.Logger)
			archiveReader := relProvider.NewExtractingArchiveReader()

			release, err := archiveReader.Read(targetTarball.Name())
			Expect(err).ToNot(HaveOccurred())

			defer release.CleanUp() //nolint:errcheck

			pkg1 := release.Packages()[0]
			Expect(fs.ReadFileString(filepath.Join(pkg1.ExtractedPath(), "in-src"))).To(Equal("in-src"))
		})

		By("updating content in the upstream release's `pkg1`", func() {
			err := fs.WriteFileString(filepath.Join(upstreamDir, "src", "in-src"), "in-src-updated")
			Expect(err).ToNot(HaveOccurred())
		})

		By("adding a package `dependent-pkg` to the upstream release", func() {
			createAndExecCommand(cmdFactory, []string{"generate-package", "dependent-pkg", "--dir", upstreamDir})

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
		})

		By("making the upstream release's package `pkg1` dependent on `dependent-pkg`", func() {
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

			createAndExecCommand(cmdFactory, []string{"create-release", "--final", "--force", "--dir", upstreamDir})
		})

		By("again running `vendor-package` to vendor the upstream release's package `pkg1`", func() {
			createAndExecCommand(cmdFactory, []string{"vendor-package", "pkg1", upstreamDir, "--dir", targetDir})
		})

		By("verifying that both `pkg1` and its dependency `dependent-pkg` have both been vendored", func() {
			targetTarball, err := fs.TempFile("bosh-vendor-package-int-test")
			Expect(err).ToNot(HaveOccurred())

			defer fs.RemoveAll(targetTarball.Name()) //nolint:errcheck

			createAndExecCommand(cmdFactory, []string{"create-release", "--tarball", targetTarball.Name(), "--force", "--dir", targetDir})

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
		})
	})
})
