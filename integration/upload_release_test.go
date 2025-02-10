package integration_test

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	boshsys "github.com/cloudfoundry/bosh-utils/system"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
)

var _ = Describe("upload-release command", func() {
	It("can upload release via git protocol", func() {
		tmpDir, err := fs.TempDir("bosh-upload-release-int-test")
		Expect(err).ToNot(HaveOccurred())

		defer fs.RemoveAll(tmpDir) //nolint:errcheck

		relName := filepath.Base(tmpDir)

		By("running `init-release`, `generate-job`, and `generate-package`", func() {
			createAndExecCommand(cmdFactory, []string{"init-release", "--git", "--dir", tmpDir})
			createAndExecCommand(cmdFactory, []string{"generate-job", "job1", "--dir", tmpDir})
			createAndExecCommand(cmdFactory, []string{"generate-package", "pkg1", "--dir", tmpDir})
		})

		By("creating a job that depends on `pkg1`", func() {
			jobSpecPath := filepath.Join(tmpDir, "jobs", "job1", "spec")

			contents, err := fs.ReadFileString(jobSpecPath)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(jobSpecPath, strings.Replace(contents, "packages: []", "packages: [pkg1]", -1))
			Expect(err).ToNot(HaveOccurred())
		})

		By("adding some content", func() {
			err := fs.WriteFileString(filepath.Join(tmpDir, "src", "in-src"), "in-src")
			Expect(err).ToNot(HaveOccurred())

			pkg1SpecPath := filepath.Join(tmpDir, "packages", "pkg1", "spec")

			contents, err := fs.ReadFileString(pkg1SpecPath)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(pkg1SpecPath, strings.Replace(contents, "files: []", "files:\n- in-src", -1))
			Expect(err).ToNot(HaveOccurred())
		})

		By("creating a release with local blobstore", func() {
			blobstoreDir := filepath.Join(tmpDir, ".blobstore")

			err := fs.MkdirAll(blobstoreDir, 0777)
			Expect(err).ToNot(HaveOccurred())

			finalYaml := "name: " + relName + `
blobstore:
  provider: local
  options:
    blobstore_path: ` + blobstoreDir

			err = fs.WriteFileString(filepath.Join(tmpDir, "config", "final.yml"), finalYaml)
			Expect(err).ToNot(HaveOccurred())

			execGit := func(args []string) {
				cmd := boshsys.Command{
					Name:       "git",
					Args:       args,
					WorkingDir: tmpDir, // --git-dir/--work-tree/etc. dont work great
				}
				_, _, _, err := deps.CmdRunner.RunComplexCommand(cmd)
				Expect(err).ToNot(HaveOccurred())
			}

			execGit([]string{"config", "--local", "user.email", "bosh-upload-release-int-test"})
			execGit([]string{"config", "--local", "user.name", "bosh-upload-release-int-test"})

			execGit([]string{"add", "-A"})
			execGit([]string{"commit", "-m", "init"})

			createAndExecCommand(cmdFactory, []string{"create-release", "--dir", tmpDir, "--final"})

			execGit([]string{"add", "-A"})
			execGit([]string{"commit", "-m", "Final release 1"})
		})

		uploadedReleaseFile := filepath.Join(tmpDir, "release-3.tgz")

		By("mocking the director's HTTP interface", func() {
			directorCACert, director := buildHTTPSServer()
			defer director.Close()

			director.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/info"),
					ghttp.RespondWith(http.StatusOK, `{"user_authentication":{"type":"basic","options":{}}}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/releases"),
					ghttp.RespondWith(http.StatusOK, "[]"),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/packages/matches"),
					ghttp.RespondWith(http.StatusOK, "[]"),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/releases"),
					ghttp.RespondWith(http.StatusOK, `{"id":123, "state":"done"}`),
					func(w http.ResponseWriter, req *http.Request) {
						defer req.Body.Close() //nolint:errcheck

						body, err := io.ReadAll(req.Body)
						Expect(err).ToNot(HaveOccurred())

						err = fs.WriteFile(uploadedReleaseFile, body)
						Expect(err).ToNot(HaveOccurred())
					},
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/tasks/123"),
					ghttp.RespondWith(http.StatusOK, `{"id":123, "state":"done"}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/tasks/123/output", "type=event"),
					ghttp.RespondWith(http.StatusOK, ``),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/tasks/123/output", "type=result"),
					ghttp.RespondWith(http.StatusOK, ``),
				),
			)

			createAndExecCommand(cmdFactory, []string{"upload-release", "git+file://" + tmpDir, "-e", director.URL(), "--ca-cert", directorCACert})
		})

		By("checking the contents of the uploaded release", func() {
			relProvider := boshrel.NewProvider(deps.CmdRunner, deps.Compressor, deps.DigestCalculator, deps.FS, deps.Logger)
			archiveReader := relProvider.NewExtractingArchiveReader()

			release, err := archiveReader.Read(uploadedReleaseFile)
			Expect(err).ToNot(HaveOccurred())

			defer release.CleanUp() //nolint:errcheck

			pkg1 := release.Packages()[0]
			Expect(fs.ReadFileString(filepath.Join(pkg1.ExtractedPath(), "in-src"))).To(Equal("in-src"))
		})
	})

	It("can upload release tarball", func() {
		boshTmpDir := filepath.Join(testHome, ".bosh", "tmp")

		tmpDir, err := fs.TempDir("bosh-upload-release-int-test")
		Expect(err).ToNot(HaveOccurred())

		defer fs.RemoveAll(tmpDir) //nolint:errcheck

		By("running `init-release`, `generate-job`, and `generate-package`", func() {
			createAndExecCommand(cmdFactory, []string{"init-release", "--dir", tmpDir})
			createAndExecCommand(cmdFactory, []string{"generate-job", "job1", "--dir", tmpDir})
			createAndExecCommand(cmdFactory, []string{"generate-package", "pkg1", "--dir", tmpDir})
		})

		By("creating a job that depends on `pkg1`", func() {
			jobSpecPath := filepath.Join(tmpDir, "jobs", "job1", "spec")

			contents, err := fs.ReadFileString(jobSpecPath)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(jobSpecPath, strings.Replace(contents, "packages: []", "packages: [pkg1]", -1))
			Expect(err).ToNot(HaveOccurred())
		})

		By("adding some content", func() {
			err := fs.WriteFileString(filepath.Join(tmpDir, "src", "in-src"), "in-src")
			Expect(err).ToNot(HaveOccurred())

			pkg1SpecPath := filepath.Join(tmpDir, "packages", "pkg1", "spec")

			contents, err := fs.ReadFileString(pkg1SpecPath)
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString(pkg1SpecPath, strings.Replace(contents, "files: []", "files:\n- in-src", -1))
			Expect(err).ToNot(HaveOccurred())
		})

		releaseTarballFile := filepath.Join(tmpDir, "release-tarball.tgz")

		By("creating a dev release", func() {
			createAndExecCommand(cmdFactory, []string{"create-release", "--dir", tmpDir, "--tarball", releaseTarballFile})
		})

		By("starting with an empty bosh tmpdir", func() {
			err := fs.RemoveAll(boshTmpDir)
			Expect(err).ToNot(HaveOccurred())
		})

		uploadedReleaseFile := filepath.Join(tmpDir, "release-3.tgz")

		By("mocking the director's HTTP interface", func() {
			directorCACert, director := buildHTTPSServer()
			defer director.Close()

			director.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/info"),
					ghttp.RespondWith(http.StatusOK, `{"user_authentication":{"type":"basic","options":{}}}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/releases"),
					ghttp.RespondWith(http.StatusOK, "[]"),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/packages/matches"),
					ghttp.RespondWith(http.StatusOK, "[]"),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/releases"),
					ghttp.RespondWith(http.StatusOK, `{"id":123, "state":"done"}`),
					func(w http.ResponseWriter, req *http.Request) {
						defer req.Body.Close() //nolint:errcheck

						body, err := io.ReadAll(req.Body)
						Expect(err).ToNot(HaveOccurred())

						err = fs.WriteFile(uploadedReleaseFile, body)
						Expect(err).ToNot(HaveOccurred())
					},
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/tasks/123"),
					ghttp.RespondWith(http.StatusOK, `{"id":123, "state":"done"}`),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/tasks/123/output", "type=event"),
					ghttp.RespondWith(http.StatusOK, ``),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/tasks/123/output", "type=result"),
					ghttp.RespondWith(http.StatusOK, ``),
				),
			)

			createAndExecCommand(cmdFactory, []string{"upload-release", releaseTarballFile, "-e", director.URL(), "--ca-cert", directorCACert})
		})

		By("checking the contents of the uploaded release", func() {
			relProvider := boshrel.NewProvider(deps.CmdRunner, deps.Compressor, deps.DigestCalculator, deps.FS, deps.Logger)
			archiveReader := relProvider.NewExtractingArchiveReader()

			release, err := archiveReader.Read(uploadedReleaseFile)
			Expect(err).ToNot(HaveOccurred())

			defer release.CleanUp() //nolint:errcheck

			pkg1 := release.Packages()[0]
			Expect(fs.ReadFileString(filepath.Join(pkg1.ExtractedPath(), "in-src"))).To(Equal("in-src"))

			err = release.CleanUp()
			Expect(err).ToNot(HaveOccurred())
		})

		err = fs.RemoveAll(tmpDir)
		Expect(err).ToNot(HaveOccurred())

		By("expecting the bosh tmpdir to be empty we can detect file leakage", func() {
			matches, err := fs.RecursiveGlob(filepath.Join(boshTmpDir, "*"))
			Expect(err).ToNot(HaveOccurred())
			Expect(matches).To(BeEmpty())
		})
	})
})
