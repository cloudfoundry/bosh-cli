package integration_test

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

var _ = Describe("finalize-release command", func() {
	var releaseDir string

	releaseName := "test-release"

	Context("when finalizing a release that was built elsewhere", func() {

		BeforeEach(func() {
			tmpDir, err := fs.TempDir("bosh-finalize-release-int-test")
			Expect(err).ToNot(HaveOccurred())
			releaseDir = tmpDir // to use the releaseDir value in other functions

			DeferCleanup(func() {
				err = fs.RemoveAll(tmpDir)
				Expect(err).ToNot(HaveOccurred())
			})

			By("running `init-release`, `generate-job`, and `generate-package`", func() {
				createAndExecCommand(cmdFactory, []string{"init-release", "--git", "--dir", releaseDir})
				createAndExecCommand(cmdFactory, []string{"generate-job", "job1", "--dir", releaseDir})
				createAndExecCommand(cmdFactory, []string{"generate-package", "pkg1", "--dir", releaseDir})
			})

			By("creating a job that depends on `pkg1`", func() {
				jobSpecPath := filepath.Join(releaseDir, "jobs", "job1", "spec")

				contents, err := fs.ReadFileString(jobSpecPath)
				Expect(err).ToNot(HaveOccurred())

				err = fs.WriteFileString(jobSpecPath, strings.Replace(contents, "packages: []", "packages: [pkg1]", -1))
				Expect(err).ToNot(HaveOccurred())
			})

			By("adding some content", func() {
				err := fs.WriteFileString(filepath.Join(releaseDir, "src", "in-src"), "in-src")
				Expect(err).ToNot(HaveOccurred())

				pkg1SpecPath := filepath.Join(releaseDir, "packages", "pkg1", "spec")

				contents, err := fs.ReadFileString(pkg1SpecPath)
				Expect(err).ToNot(HaveOccurred())

				err = fs.WriteFileString(pkg1SpecPath, strings.Replace(contents, "files: []", "files:\n- in-src", -1))
				Expect(err).ToNot(HaveOccurred())
			})

			By("creating a release with local blobstore", func() {
				blobstoreDir := filepath.Join(releaseDir, ".blobstore")

				err := fs.MkdirAll(blobstoreDir, 0777)
				Expect(err).ToNot(HaveOccurred())

				finalYaml := "name: " + releaseName + `
blobstore:
  provider: local
  options:
    blobstore_path: ` + blobstoreDir

				err = fs.WriteFileString(filepath.Join(releaseDir, "config", "final.yml"), finalYaml)
				Expect(err).ToNot(HaveOccurred())
			})

			createAndExecCommand(cmdFactory, []string{"create-release", "--dir", releaseDir, fmt.Sprintf("--tarball=%s/release.tgz", releaseDir), "--force"})

		})

		It("updates the .final_builds index for each job and package", func() {
			createAndExecCommand(cmdFactory, []string{"finalize-release", "--dir", releaseDir, fmt.Sprintf("%s/release.tgz", releaseDir), "--force"})

			jobContents, err := fs.ReadFileString(fmt.Sprintf("%s/.final_builds/jobs/job1/index.yml", releaseDir))
			Expect(err).ToNot(HaveOccurred())
			Expect(jobContents).ToNot(BeEmpty())

			type JobContents struct {
				Builds struct {
					Id struct {
						Version     string `yaml:"version"`
						BlobstoreID string `yaml:"blobstore_id"`
						Sha1        string `yaml:"sha1"`
					} `yaml:"7225651667f52a2c600a4b7271c76b7277268730574d55dae509c0e9ad6c89e7"`
				} `yaml:"builds"`
				FormatVersion string `yaml:"format-version"`
			}
			jobs := JobContents{}
			err = yaml.Unmarshal([]byte(jobContents), &jobs)
			Expect(err).ToNot(HaveOccurred())

			Expect(jobs.Builds.Id.Version).To(Equal("7225651667f52a2c600a4b7271c76b7277268730574d55dae509c0e9ad6c89e7"))
			Expect(jobs.Builds.Id.BlobstoreID).NotTo(BeEmpty())
			Expect(jobs.Builds.Id.Sha1).NotTo(BeEmpty())
			Expect(jobs.FormatVersion).To(Equal("2"))

			pkgContents, err := fs.ReadFileString(fmt.Sprintf("%s/.final_builds/packages/pkg1/index.yml", releaseDir))
			Expect(err).ToNot(HaveOccurred())
			Expect(jobContents).ToNot(BeEmpty())

			type PackageContents struct {
				Builds struct {
					Id struct {
						Version     string `yaml:"version"`
						BlobstoreID string `yaml:"blobstore_id"`
						Sha1        string `yaml:"sha1"`
					} `yaml:"074b9ff2fd95d50d32db373ae16bd8cb5d6e098beb7ff35745ea1b5115264710"`
				} `yaml:"builds"`
				FormatVersion string `yaml:"format-version"`
			}

			packages := PackageContents{}
			err = yaml.Unmarshal([]byte(pkgContents), &packages)
			Expect(err).ToNot(HaveOccurred())

			Expect(packages.Builds.Id.Version).To(Equal("074b9ff2fd95d50d32db373ae16bd8cb5d6e098beb7ff35745ea1b5115264710"))
			Expect(packages.Builds.Id.BlobstoreID).NotTo(BeEmpty())
			Expect(packages.Builds.Id.Sha1).NotTo(BeEmpty())
			Expect(packages.FormatVersion).To(Equal("2"))
		})

		It("prints release summary", func() {
			createAndExecCommand(cmdFactory, []string{"finalize-release", "--dir", releaseDir, fmt.Sprintf("%s/release.tgz", releaseDir), "--force"})
			output := strings.Join(ui.Said, " ")
			Expect(output).To(ContainSubstring("Added job 'job1/7225651667f52a2c600a4b7271c76b7277268730574d55dae509c0e9ad6c89e7'"))
			Expect(output).To(ContainSubstring("Added package 'pkg1/074b9ff2fd95d50d32db373ae16bd8cb5d6e098beb7ff35745ea1b5115264710'"))
			Expect(output).To(ContainSubstring("Added final release 'test-release/1'"))
		})

		It("cannot create a final release without the blobstore configured", func() {
			err := fs.WriteFileString(filepath.Join(releaseDir, "config", "final.yml"), "")
			Expect(err).ToNot(HaveOccurred())

			command := createCommand(cmdFactory, []string{"finalize-release", "--dir", releaseDir, fmt.Sprintf("%s/release.tgz", releaseDir), "--force"})
			err = command.Execute()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("Expected non-empty 'blobstore.provider' in config .*/config/final\\.yml"))
		})

		It("cannot create a final release without the blobstore secret configured", func() {
			finalYaml := `---
blobstore:
  provider: s3
  options:
    bucket_name: test
`
			err := fs.WriteFileString(filepath.Join(releaseDir, "config", "final.yml"), finalYaml)
			Expect(err).ToNot(HaveOccurred())

			command := createCommand(cmdFactory, []string{"finalize-release", "--dir", releaseDir, fmt.Sprintf("%s/release.tgz", releaseDir), "--force"})
			err = command.Execute()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(MatchRegexp("Creating blob in inner blobstore: Generating blobstore ID: the client operates in read only mode. Change 'credentials_source' parameter value"))
		})

		Context("when no previous releases have been made", func() {
			It("finalize-release uploads the job & package blobs", func() {
				Expect(fs.FileExists(fmt.Sprintf("%s/releases", releaseDir))).To(Equal(false))
				//Expect(fs.FileExists(fmt.Sprintf("%s/dev_releases", releaseDir))).To(Equal(false))
				Expect(fs.FileExists(fmt.Sprintf("%s/.final_releases", releaseDir))).To(Equal(false))
				// Expect(fs.FileExists(fmt.Sprintf("%s/.dev_builds", releaseDir))).To(Equal(false))
				// Expect(fs.FileExists(fmt.Sprintf("%s/.blobstore", releaseDir))).To(Equal(false))

				createAndExecCommand(cmdFactory, []string{"finalize-release", "--dir", releaseDir, fmt.Sprintf("%s/release.tgz", releaseDir), "--force"})

				Expect(fs.FileExists(fmt.Sprintf("%s/releases/%s/%s-1.yml", releaseDir, releaseName, releaseName))).To(Equal(true))
				Expect(fs.FileExists(fmt.Sprintf("%s/.final_builds/jobs/job1/index.yml", releaseDir))).To(Equal(true))
				Expect(fs.FileExists(fmt.Sprintf("%s/.final_builds/packages/pkg1/index.yml", releaseDir))).To(Equal(true))
				//	Dir.chdir(IntegrationSupport::ClientSandbox.test_release_dir) do
				//	  expect(Dir).to_not exist('releases')
				//	  expect(Dir).to_not exist('dev_releases')
				//	  expect(Dir).to_not exist('.final_builds')
				//	  expect(Dir).to_not exist('.dev_builds')
				//	  expect(Dir).to_not exist(IntegrationSupport::ClientSandbox.blobstore_dir)
				//
				//	  bosh_runner.run_in_current_dir("finalize-release #{asset_path('dummy-gocli-release.tgz')} --force")
				//	  expect(File).to exist('releases/dummy/dummy-1.yml')
				//	  expect(File).to exist('.final_builds/jobs/dummy/index.yml')
				//	  expect(File).to exist('.final_builds/packages/bad_package/index.yml')
				//	  uploaded_blob_count = Dir[File.join(IntegrationSupport::ClientSandbox.blobstore_dir, '**', '*')].length
				//	  expect(uploaded_blob_count).to eq(7)
				//	end
			})
		})
	})

	Context("when finalizing a release that was built in the current release dir", func() {

		BeforeEach(func() {
			tmpDir, err := fs.TempDir("bosh-finalize-release-int-test")
			Expect(err).ToNot(HaveOccurred())
			releaseDir = tmpDir // to use the releaseDir value in other functions

			DeferCleanup(func() {
				err = fs.RemoveAll(tmpDir)
				Expect(err).ToNot(HaveOccurred())
			})

			By("running `init-release`, `generate-job`, and `generate-package`", func() {
				createAndExecCommand(cmdFactory, []string{"init-release", "--git", "--dir", releaseDir})
				createAndExecCommand(cmdFactory, []string{"generate-job", "job1", "--dir", releaseDir})
				createAndExecCommand(cmdFactory, []string{"generate-package", "pkg1", "--dir", releaseDir})
			})

			By("creating a job that depends on `pkg1`", func() {
				jobSpecPath := filepath.Join(releaseDir, "jobs", "job1", "spec")

				contents, err := fs.ReadFileString(jobSpecPath)
				Expect(err).ToNot(HaveOccurred())

				err = fs.WriteFileString(jobSpecPath, strings.Replace(contents, "packages: []", "packages: [pkg1]", -1))
				Expect(err).ToNot(HaveOccurred())
			})

			By("adding some content", func() {
				err := fs.WriteFileString(filepath.Join(releaseDir, "src", "in-src"), "in-src")
				Expect(err).ToNot(HaveOccurred())

				pkg1SpecPath := filepath.Join(releaseDir, "packages", "pkg1", "spec")

				contents, err := fs.ReadFileString(pkg1SpecPath)
				Expect(err).ToNot(HaveOccurred())

				err = fs.WriteFileString(pkg1SpecPath, strings.Replace(contents, "files: []", "files:\n- in-src", -1))
				Expect(err).ToNot(HaveOccurred())
			})

			By("creating a release with local blobstore", func() {
				blobstoreDir := filepath.Join(releaseDir, ".blobstore")

				err := fs.MkdirAll(blobstoreDir, 0777)
				Expect(err).ToNot(HaveOccurred())

				finalYaml := "name: " + releaseName + `
blobstore:
  provider: local
  options:
    blobstore_path: ` + blobstoreDir

				err = fs.WriteFileString(filepath.Join(releaseDir, "config", "final.yml"), finalYaml)
				Expect(err).ToNot(HaveOccurred())
			})

			err = fs.WriteFileString(filepath.Join(releaseDir, "LICENSE"), "")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(filepath.Join(releaseDir, "NOTICE"), "")
			Expect(err).ToNot(HaveOccurred())
		})

		It("can finalize the dev release tarball", func() {
			err := os.Chdir(releaseDir)
			Expect(err).ToNot(HaveOccurred())

			os.Remove("LICENSE")
			Expect(fs.FileExists("LICENSE")).To(Equal(false))
			os.Remove("NOTICE")
			Expect(fs.FileExists("NOTICE")).To(Equal(false))

			createAndExecCommand(cmdFactory, []string{"create-release", "--name", "test-release", fmt.Sprintf("--tarball=%s/release.tgz", releaseDir), "--force"})

			createAndExecCommand(cmdFactory, []string{"finalize-release", fmt.Sprintf("%s/release.tgz", releaseDir), "--force"})
		})

		It("works without a NOTICE or LICENSE present", func() {
			err := os.Chdir(releaseDir)
			Expect(err).ToNot(HaveOccurred())

			os.Remove("LICENSE")
			Expect(fs.FileExists("LICENSE")).To(Equal(false))
			os.Remove("NOTICE")
			Expect(fs.FileExists("NOTICE")).To(Equal(false))

			createAndExecCommand(cmdFactory, []string{"create-release", "--name", "test-release", fmt.Sprintf("--tarball=%s/release.tgz", releaseDir), "--force"})

			createAndExecCommand(cmdFactory, []string{"finalize-release", fmt.Sprintf("%s/release.tgz", releaseDir), "--force"})
			output := strings.Join(ui.Said, " ")
			Expect(output).ToNot(ContainSubstring("Added license"))
		})

		It("includes the LICENSE file", func() {
			err := os.Chdir(releaseDir)
			Expect(err).ToNot(HaveOccurred())

			os.Remove("NOTICE")
			Expect(fs.FileExists("NOTICE")).To(Equal(false))
			err = fs.WriteFileString(filepath.Join(releaseDir, "LICENSE"), "This is an example license file")
			Expect(err).ToNot(HaveOccurred())

			createAndExecCommand(cmdFactory, []string{"create-release", "--name", "test-release", fmt.Sprintf("--tarball=%s/release.tgz", releaseDir), "--force"})

			createAndExecCommand(cmdFactory, []string{"finalize-release", fmt.Sprintf("%s/release.tgz", releaseDir), "--force"})
			output := strings.Join(ui.Said, " ")
			expectedLicenseVersion := "24f59b89c3a9f4eed2f4b9b07bc754891fadc49d8ec0dda25c562d90e568b375"
			Expect(output).To(ContainSubstring("Added license 'license/24f59b89c3a9f4eed2f4b9b07bc754891fadc49d8ec0dda25c562d90e568b375"))

			fs.FileExists(releaseDir + expectedLicenseVersion)
			releaseTarball := listTarballContents(fmt.Sprintf("%s/release.tgz", releaseDir))
			Expect(releaseTarball).To(ContainElement("LICENSE"))

			verifyDigest(releaseDir, expectedLicenseVersion)
		})

		It("includes the NOTICE file if no LICENSE was present", func() {
			err := os.Chdir(releaseDir)
			Expect(err).ToNot(HaveOccurred())

			os.Remove("LICENSE")
			Expect(fs.FileExists("LICENSE")).To(Equal(false))
			err = fs.WriteFileString(filepath.Join(releaseDir, "NOTICE"), "This is an example license file called NOTICE")
			Expect(err).ToNot(HaveOccurred())

			createAndExecCommand(cmdFactory, []string{"create-release", "--name", "test-release", fmt.Sprintf("--tarball=%s/release.tgz", releaseDir), "--force"})

			createAndExecCommand(cmdFactory, []string{"finalize-release", fmt.Sprintf("%s/release.tgz", releaseDir), "--force"})
			output := strings.Join(ui.Said, " ")
			expectedLicenseVersion := "eb70cb1dcc90b1d1b1271bfa26a57e62240e46a38a87c64fbde94e2b65fe0c37"
			Expect(output).To(ContainSubstring("Added license 'license/eb70cb1dcc90b1d1b1271bfa26a57e62240e46a38a87c64fbde94e2b65fe0c37"))

			fs.FileExists(releaseDir + expectedLicenseVersion)
			releaseTarball := listTarballContents(fmt.Sprintf("%s/release.tgz", releaseDir))
			Expect(releaseTarball).To(ContainElement("NOTICE"))

			verifyDigest(releaseDir, expectedLicenseVersion)
		})

		It("includes both NOTICE and LICENSE files when present", func() {
			err := os.Chdir(releaseDir)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(filepath.Join(releaseDir, "NOTICE"), "This is an example license file called NOTICE")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString(filepath.Join(releaseDir, "LICENSE"), "This is an example license file")
			Expect(err).ToNot(HaveOccurred())

			createAndExecCommand(cmdFactory, []string{"create-release", "--name", "test-release", fmt.Sprintf("--tarball=%s/release.tgz", releaseDir), "--force"})

			createAndExecCommand(cmdFactory, []string{"finalize-release", fmt.Sprintf("%s/release.tgz", releaseDir), "--force"})
			output := strings.Join(ui.Said, " ")
			expectedLicenseVersion := "9f6d8d782d57bcca93b645a0cbd0a95b943c72bf61093e1874aff6b8b54c371e"
			Expect(output).To(ContainSubstring("Added license 'license/9f6d8d782d57bcca93b645a0cbd0a95b943c72bf61093e1874aff6b8b54c371e"))

			fs.FileExists(releaseDir + expectedLicenseVersion)
			releaseTarball := listTarballContents(fmt.Sprintf("%s/release.tgz", releaseDir))
			Expect(releaseTarball).To(ContainElement("NOTICE"))
			Expect(releaseTarball).To(ContainElement("LICENSE"))

			verifyDigest(releaseDir, expectedLicenseVersion)
		})
	})
})

func verifyDigest(releasedir, expectedLicenseVersion string) {
	type LicenseIndex struct {
		Builds map[string]struct {
			BlobstoreID string `yaml:"blobstore_id"`
			Sha1        string `yaml:"sha1"`
		} `yaml:"builds"`
	}
	finalBuildsLicenseFile, err := os.ReadFile(fmt.Sprintf("%s/.final_builds/license/index.yml", releasedir))
	Expect(err).ToNot(HaveOccurred())

	var licenseIndex LicenseIndex
	err = yaml.Unmarshal(finalBuildsLicenseFile, &licenseIndex)
	Expect(err).ToNot(HaveOccurred())
	blobstoreID := licenseIndex.Builds[expectedLicenseVersion].BlobstoreID

	licenseFile, err := os.Open(fmt.Sprintf("%s/.blobstore/%s", releasedir, blobstoreID))
	Expect(err).ToNot(HaveOccurred())
	hash := sha256.New()
	_, err = io.Copy(hash, licenseFile)
	Expect(err).ToNot(HaveOccurred())
	actualDigest := fmt.Sprintf("%x", hash.Sum(nil))

	expectedDigest := licenseIndex.Builds[expectedLicenseVersion].Sha1
	Expect(actualDigest).To(Equal(strings.Split(expectedDigest, ":")[1]))
}
