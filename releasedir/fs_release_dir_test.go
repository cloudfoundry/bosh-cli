package releasedir_test

import (
	"errors"
	"os"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/clock/fakeclock"
	fakesys "github.com/cloudfoundry/bosh-utils/system/fakes"
	semver "github.com/cppforlife/go-semi-semantic/version"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	boshrel "github.com/cloudfoundry/bosh-cli/v7/release"
	boshman "github.com/cloudfoundry/bosh-cli/v7/release/manifest"
	boshpkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	fakerel "github.com/cloudfoundry/bosh-cli/v7/release/releasefakes"
	"github.com/cloudfoundry/bosh-cli/v7/release/resource"
	fakeres "github.com/cloudfoundry/bosh-cli/v7/release/resource/resourcefakes"
	. "github.com/cloudfoundry/bosh-cli/v7/releasedir"
	fakereldir "github.com/cloudfoundry/bosh-cli/v7/releasedir/releasedirfakes"
)

type failingPackageGenerator struct {
	FSGenerator
	err error
}

func (g failingPackageGenerator) GeneratePackage(GeneratePackageOptions) error {
	return g.err
}

type packagingFailureGenerator struct {
	FSGenerator
	fs    *fakesys.FakeFileSystem
	armed bool
}

func (g *packagingFailureGenerator) GeneratePackage(options GeneratePackageOptions) error {
	if !g.armed {
		delete(g.fs.WriteFileErrors, "/dir/packages/sample-pkg/packaging")
	}

	err := g.FSGenerator.GeneratePackage(options)
	if err != nil {
		return err
	}

	if g.armed {
		g.fs.WriteFileErrors["/dir/packages/sample-pkg/packaging"] = errors.New("fake-err")
	}

	return nil
}

var _ = Describe("FSGenerator", func() {
	var (
		config        *fakereldir.FakeConfig
		gitRepo       *fakereldir.FakeGitRepo
		blobsDir      *fakereldir.FakeBlobsDir
		gen           *fakereldir.FakeGenerator
		devReleases   *fakereldir.FakeReleaseIndex
		finalReleases *fakereldir.FakeReleaseIndex
		finalIndicies boshrel.ArchiveIndicies
		reader        *fakerel.FakeReader
		timeService   clock.Clock
		fs            *fakesys.FakeFileSystem
		releaseDir    FSReleaseDir
	)

	BeforeEach(func() {
		config = &fakereldir.FakeConfig{}
		gitRepo = &fakereldir.FakeGitRepo{}
		blobsDir = &fakereldir.FakeBlobsDir{}
		gen = &fakereldir.FakeGenerator{}
		devReleases = &fakereldir.FakeReleaseIndex{}
		finalReleases = &fakereldir.FakeReleaseIndex{}
		finalIndicies = boshrel.ArchiveIndicies{
			Jobs:     &fakeres.FakeArchiveIndex{},
			Packages: &fakeres.FakeArchiveIndex{},
		}
		reader = &fakerel.FakeReader{}
		timeService = fakeclock.NewFakeClock(time.Date(2009, time.November, 10, 23, 1, 2, 333, time.UTC))
		fs = fakesys.NewFakeFileSystem()
		releaseDir = NewFSReleaseDir(
			"/dir",
			config,
			gitRepo,
			blobsDir,
			gen,
			devReleases,
			finalReleases,
			finalIndicies,
			reader,
			timeService,
			fs,
			2,
		)
	})

	Describe("Init", func() {
		It("creates common jobs, packages and src directories", func() {
			err := releaseDir.Init(InitOptions{Git: true})
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.FileExists("/dir/jobs")).To(BeTrue())
			Expect(fs.FileExists("/dir/packages")).To(BeTrue())
			Expect(fs.FileExists("/dir/src")).To(BeTrue())
		})

		It("returns error if creating common dirs fails", func() {
			fs.MkdirAllError = errors.New("fake-err")

			err := releaseDir.Init(InitOptions{Git: true})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("saves release name to directory base name", func() {
			err := releaseDir.Init(InitOptions{Git: true})
			Expect(err).ToNot(HaveOccurred())

			Expect(config.SaveNameCallCount()).To(Equal(1))
			Expect(config.SaveNameArgsForCall(0)).To(Equal("dir"))
		})

		It("saves an explicit release name", func() {
			err := releaseDir.Init(InitOptions{Name: "my-release"})
			Expect(err).ToNot(HaveOccurred())

			Expect(config.SaveNameArgsForCall(0)).To(Equal("my-release"))
		})

		It("keeps an existing release name when reinitializing", func() {
			config.NameReturns("existing-name", nil)

			err := releaseDir.Init(InitOptions{Force: true})
			Expect(err).ToNot(HaveOccurred())

			Expect(config.SaveNameArgsForCall(0)).To(Equal("existing-name"))
		})

		It("returns error if reading the existing release name fails", func() {
			config.NameReturns("", errors.New("fake-err"))

			err := releaseDir.Init(InitOptions{Force: true})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
			Expect(config.SaveNameCallCount()).To(Equal(0))
		})

		It("saves release name to directory base name stripping '-release' suffix from the name", func() {
			releaseDir := NewFSReleaseDir(
				"/dir-release",
				config,
				gitRepo,
				blobsDir,
				gen,
				devReleases,
				finalReleases,
				finalIndicies,
				reader,
				timeService,
				fs,
				2,
			)

			err := releaseDir.Init(InitOptions{Git: true})
			Expect(err).ToNot(HaveOccurred())

			Expect(config.SaveNameCallCount()).To(Equal(1))
			Expect(config.SaveNameArgsForCall(0)).To(Equal("dir"))
		})

		It("returns error if saving final name fails", func() {
			config.SaveNameReturns(errors.New("fake-err"))

			err := releaseDir.Init(InitOptions{Git: true})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("inits blobs when blobs.yml does not exist", func() {
			err := releaseDir.Init(InitOptions{Git: true})
			Expect(err).ToNot(HaveOccurred())

			Expect(blobsDir.InitCallCount()).To(Equal(1))
		})

		It("does not overwrite an existing blobs.yml when forced", func() {
			err := fs.WriteFileString("/dir/config/blobs.yml", "--- {}\n")
			Expect(err).ToNot(HaveOccurred())

			err = releaseDir.Init(InitOptions{Force: true})
			Expect(err).ToNot(HaveOccurred())

			Expect(blobsDir.InitCallCount()).To(Equal(0))
		})

		It("returns error if initing blobs fails", func() {
			blobsDir.InitReturns(errors.New("fake-err"))

			err := releaseDir.Init(InitOptions{Git: true})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("writes .gitignore even when git is not requested", func() {
			err := releaseDir.Init(InitOptions{})
			Expect(err).ToNot(HaveOccurred())

			Expect(gitRepo.WriteIgnoreFileCallCount()).To(Equal(1))
			Expect(gitRepo.InitCallCount()).To(Equal(0))
		})

		It("inits git repo if requested", func() {
			err := releaseDir.Init(InitOptions{Git: true})
			Expect(err).ToNot(HaveOccurred())

			Expect(gitRepo.InitCallCount()).To(Equal(1))
		})

		It("returns error if writing .gitignore fails", func() {
			gitRepo.WriteIgnoreFileReturns(errors.New("fake-err"))

			err := releaseDir.Init(InitOptions{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if initing git repo fails", func() {
			gitRepo.InitReturns(errors.New("fake-err"))

			err := releaseDir.Init(InitOptions{Git: true})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("refuses to reinitialize an existing release without force", func() {
			err := fs.WriteFileString("/dir/config/final.yml", "name: dir\n")
			Expect(err).ToNot(HaveOccurred())

			err = releaseDir.Init(InitOptions{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("already initialized"))
			Expect(err.Error()).To(ContainSubstring("--force"))
		})

		It("reinitializes an existing release with force", func() {
			err := fs.WriteFileString("/dir/config/final.yml", "name: dir\n")
			Expect(err).ToNot(HaveOccurred())

			err = releaseDir.Init(InitOptions{Force: true})
			Expect(err).ToNot(HaveOccurred())
		})

		It("saves a configured blobstore", func() {
			err := releaseDir.Init(InitOptions{
				BlobstoreProvider:       "s3",
				BlobstoreOptions:        map[string]interface{}{"bucket_name": "bucket"},
				BlobstorePrivateOptions: map[string]interface{}{"access_key_id": "key"},
			})
			Expect(err).ToNot(HaveOccurred())

			Expect(config.SaveBlobstoreCallCount()).To(Equal(1))
			provider, options, privateOptions := config.SaveBlobstoreArgsForCall(0)
			Expect(provider).To(Equal("s3"))
			Expect(options).To(Equal(map[string]interface{}{"bucket_name": "bucket"}))
			Expect(privateOptions).To(Equal(map[string]interface{}{"access_key_id": "key"}))
		})

		It("returns error if saving the blobstore fails", func() {
			config.SaveBlobstoreReturns(errors.New("fake-err"))

			err := releaseDir.Init(InitOptions{BlobstoreProvider: "s3"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("defaults a sample release to a local blobstore", func() {
			err := releaseDir.Init(InitOptions{Sample: true})
			Expect(err).ToNot(HaveOccurred())

			provider, options, _ := config.SaveBlobstoreArgsForCall(0)
			Expect(provider).To(Equal("local"))
			Expect(options).To(Equal(map[string]interface{}{"blobstore_path": ".blobs"}))
		})

		It("returns error if reading the existing blobstore fails when generating a sample", func() {
			config.BlobstoreReturns("", nil, errors.New("fake-err"))

			err := releaseDir.Init(InitOptions{Sample: true})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
			Expect(config.SaveBlobstoreCallCount()).To(Equal(0))
		})

		It("keeps an existing blobstore when generating a sample", func() {
			config.BlobstoreReturns("s3", map[string]interface{}{"bucket_name": "bucket"}, nil)

			err := releaseDir.Init(InitOptions{Sample: true, Force: true})
			Expect(err).ToNot(HaveOccurred())

			provider, options, _ := config.SaveBlobstoreArgsForCall(0)
			Expect(provider).To(Equal("s3"))
			Expect(options).To(BeNil())
		})

		It("preserves an existing blobstore when reinitializing a sample", func() {
			err := fs.WriteFileString("/dir/config/final.yml",
				"name: dir\nblobstore:\n  provider: s3\n  options:\n    bucket_name: bucket\n")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString("/dir/config/private.yml",
				"blobstore:\n  options:\n    access_key_id: secret\n")
			Expect(err).ToNot(HaveOccurred())

			releaseDir := NewFSReleaseDir(
				"/dir",
				NewFSConfig("/dir/config/final.yml", "/dir/config/private.yml", fs),
				gitRepo,
				blobsDir,
				gen,
				devReleases,
				finalReleases,
				finalIndicies,
				reader,
				timeService,
				fs,
				2,
			)

			err = releaseDir.Init(InitOptions{Sample: true, Force: true})
			Expect(err).ToNot(HaveOccurred())

			finalConfig, err := fs.ReadFileString("/dir/config/final.yml")
			Expect(err).ToNot(HaveOccurred())
			Expect(finalConfig).To(ContainSubstring("provider: s3"))
			Expect(finalConfig).To(ContainSubstring("bucket_name: bucket"))
			Expect(finalConfig).ToNot(ContainSubstring("blobstore_path"))

			privateConfig, err := fs.ReadFileString("/dir/config/private.yml")
			Expect(err).ToNot(HaveOccurred())
			Expect(privateConfig).To(ContainSubstring("access_key_id: secret"))
		})

		It("preserves a custom local blobstore path when reinitializing a sample", func() {
			err := fs.WriteFileString("/dir/config/final.yml",
				"name: dir\nblobstore:\n  provider: local\n  options:\n    blobstore_path: /custom\n")
			Expect(err).ToNot(HaveOccurred())

			releaseDir := NewFSReleaseDir(
				"/dir",
				NewFSConfig("/dir/config/final.yml", "/dir/config/private.yml", fs),
				gitRepo,
				blobsDir,
				gen,
				devReleases,
				finalReleases,
				finalIndicies,
				reader,
				timeService,
				fs,
				2,
			)

			err = releaseDir.Init(InitOptions{Sample: true, Force: true})
			Expect(err).ToNot(HaveOccurred())

			finalConfig, err := fs.ReadFileString("/dir/config/final.yml")
			Expect(err).ToNot(HaveOccurred())
			Expect(finalConfig).To(ContainSubstring("blobstore_path: /custom"))
			Expect(finalConfig).ToNot(ContainSubstring("blobstore_path: .blobs"))
		})

		It("does not mutate the caller's blobstore options", func() {
			blobstoreOptions := map[string]interface{}{"flag": "value"}

			err := releaseDir.Init(InitOptions{Sample: true, BlobstoreOptions: blobstoreOptions})
			Expect(err).ToNot(HaveOccurred())

			Expect(blobstoreOptions).To(Equal(map[string]interface{}{"flag": "value"}))
		})

		It("derives the release name when no config exists", func() {
			releaseDir := NewFSReleaseDir(
				"/my-release",
				NewFSConfig("/my-release/config/final.yml", "/my-release/config/private.yml", fs),
				gitRepo,
				blobsDir,
				gen,
				devReleases,
				finalReleases,
				finalIndicies,
				reader,
				timeService,
				fs,
				2,
			)

			err := releaseDir.Init(InitOptions{})
			Expect(err).ToNot(HaveOccurred())

			finalConfig, err := fs.ReadFileString("/my-release/config/final.yml")
			Expect(err).ToNot(HaveOccurred())
			Expect(finalConfig).To(ContainSubstring("name: my"))
		})

		It("does not overwrite the release name when the private config is malformed", func() {
			err := fs.WriteFileString("/my-release/config/final.yml", "name: configured\n")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString("/my-release/config/private.yml", "-")
			Expect(err).ToNot(HaveOccurred())

			releaseDir := NewFSReleaseDir(
				"/my-release",
				NewFSConfig("/my-release/config/final.yml", "/my-release/config/private.yml", fs),
				gitRepo,
				blobsDir,
				gen,
				devReleases,
				finalReleases,
				finalIndicies,
				reader,
				timeService,
				fs,
				2,
			)

			err = releaseDir.Init(InitOptions{Force: true})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unmarshalling config"))

			finalConfig, err := fs.ReadFileString("/my-release/config/final.yml")
			Expect(err).ToNot(HaveOccurred())
			Expect(finalConfig).To(ContainSubstring("name: configured"))
		})

		It("does not fall back to a local blobstore when the private config is malformed", func() {
			err := fs.WriteFileString("/my-release/config/final.yml",
				"name: configured\nblobstore:\n  provider: s3\n  options:\n    bucket_name: bucket\n")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString("/my-release/config/private.yml", "-")
			Expect(err).ToNot(HaveOccurred())

			releaseDir := NewFSReleaseDir(
				"/my-release",
				NewFSConfig("/my-release/config/final.yml", "/my-release/config/private.yml", fs),
				gitRepo,
				blobsDir,
				gen,
				devReleases,
				finalReleases,
				finalIndicies,
				reader,
				timeService,
				fs,
				2,
			)

			err = releaseDir.Init(InitOptions{Force: true, Sample: true, Name: "configured"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unmarshalling config"))

			finalConfig, err := fs.ReadFileString("/my-release/config/final.yml")
			Expect(err).ToNot(HaveOccurred())
			Expect(finalConfig).To(ContainSubstring("provider: s3"))
			Expect(finalConfig).ToNot(ContainSubstring("provider: local"))
		})

		It("does not rewrite the config when the private config is malformed and a name is given", func() {
			err := fs.WriteFileString("/my-release/config/final.yml",
				"name: configured\nfinal_name: legacy\nblobstore:\n  provider: s3\n")
			Expect(err).ToNot(HaveOccurred())

			err = fs.WriteFileString("/my-release/config/private.yml", "-")
			Expect(err).ToNot(HaveOccurred())

			releaseDir := NewFSReleaseDir(
				"/my-release",
				NewFSConfig("/my-release/config/final.yml", "/my-release/config/private.yml", fs),
				gitRepo,
				blobsDir,
				gen,
				devReleases,
				finalReleases,
				finalIndicies,
				reader,
				timeService,
				fs,
				2,
			)

			err = releaseDir.Init(InitOptions{Force: true, Sample: true, Name: "configured"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Unmarshalling config"))

			finalConfig, err := fs.ReadFileString("/my-release/config/final.yml")
			Expect(err).ToNot(HaveOccurred())
			Expect(finalConfig).To(ContainSubstring("name: configured"))
			Expect(finalConfig).To(ContainSubstring("final_name: legacy"))
		})

		It("repairs a config with both name and final_name when reinitializing", func() {
			err := fs.WriteFileString("/my-release/config/final.yml",
				"name: configured\nfinal_name: legacy\n")
			Expect(err).ToNot(HaveOccurred())

			releaseDir := NewFSReleaseDir(
				"/my-release",
				NewFSConfig("/my-release/config/final.yml", "/my-release/config/private.yml", fs),
				gitRepo,
				blobsDir,
				gen,
				devReleases,
				finalReleases,
				finalIndicies,
				reader,
				timeService,
				fs,
				2,
			)

			err = releaseDir.Init(InitOptions{Force: true})
			Expect(err).ToNot(HaveOccurred())

			finalConfig, err := fs.ReadFileString("/my-release/config/final.yml")
			Expect(err).ToNot(HaveOccurred())
			Expect(finalConfig).To(ContainSubstring("name: configured"))
			Expect(finalConfig).ToNot(ContainSubstring("final_name"))
		})

		It("generates a sample job, package and source file", func() {
			err := releaseDir.Init(InitOptions{Sample: true})
			Expect(err).ToNot(HaveOccurred())

			Expect(gen.GenerateJobCallCount()).To(Equal(1))
			Expect(gen.GenerateJobArgsForCall(0)).To(Equal(GenerateJobOptions{
				Name:     "sample-job",
				Packages: []string{"sample-pkg"},
			}))

			Expect(gen.GeneratePackageCallCount()).To(Equal(1))
			Expect(gen.GeneratePackageArgsForCall(0)).To(Equal(GeneratePackageOptions{
				Name:  "sample-pkg",
				Files: []string{"sample-pkg/**/*"},
			}))

			Expect(fs.ReadFileString("/dir/src/sample-pkg/file.txt")).To(Equal("Hello from sample-pkg!\n"))
			Expect(fs.ReadFileString("/dir/packages/sample-pkg/packaging")).To(Equal(
				"set -eu\n\ncp -a sample-pkg/. \"${BOSH_INSTALL_TARGET}/\"\n"))
			Expect(fs.ReadFileString("/dir/README.md")).To(ContainSubstring("# dir"))
		})

		It("writes a sample job executable that uses the sample package", func() {
			releaseDir := NewFSReleaseDir(
				"/dir",
				config,
				gitRepo,
				blobsDir,
				NewFSGenerator("/dir", fs),
				devReleases,
				finalReleases,
				finalIndicies,
				reader,
				timeService,
				fs,
				2,
			)

			err := releaseDir.Init(InitOptions{Sample: true})
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/jobs/sample-job/templates/bin/sample-job")).To(ContainSubstring(
				"cat /var/vcap/packages/sample-pkg/file.txt"))
		})

		It("returns error if generating the sample fails", func() {
			gen.GenerateJobReturns(errors.New("fake-err"))

			err := releaseDir.Init(InitOptions{Sample: true})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("removes a partially generated sample when the packaging script cannot be rewritten", func() {
			gen := &packagingFailureGenerator{
				FSGenerator: NewFSGenerator("/dir", fs),
				fs:          fs,
				armed:       true,
			}

			releaseDir := NewFSReleaseDir(
				"/dir",
				config,
				gitRepo,
				blobsDir,
				gen,
				devReleases,
				finalReleases,
				finalIndicies,
				reader,
				timeService,
				fs,
				2,
			)

			err := releaseDir.Init(InitOptions{Sample: true})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
			Expect(fs.FileExists("/dir/packages/sample-pkg")).To(BeFalse())
			Expect(fs.FileExists("/dir/jobs/sample-job")).To(BeFalse())
		})

		It("removes a partially generated sample job when package generation fails", func() {
			releaseDir := NewFSReleaseDir(
				"/dir",
				config,
				gitRepo,
				blobsDir,
				failingPackageGenerator{
					FSGenerator: NewFSGenerator("/dir", fs),
					err:         errors.New("fake-err"),
				},
				devReleases,
				finalReleases,
				finalIndicies,
				reader,
				timeService,
				fs,
				2,
			)

			err := releaseDir.Init(InitOptions{Sample: true})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
			Expect(fs.FileExists("/dir/jobs/sample-job")).To(BeFalse())
		})

		It("keeps a pre-existing sample job when package generation fails", func() {
			err := fs.MkdirAll("/dir/jobs/sample-job", os.ModePerm)
			Expect(err).ToNot(HaveOccurred())

			releaseDir := NewFSReleaseDir(
				"/dir",
				config,
				gitRepo,
				blobsDir,
				failingPackageGenerator{
					FSGenerator: NewFSGenerator("/dir", fs),
					err:         errors.New("fake-err"),
				},
				devReleases,
				finalReleases,
				finalIndicies,
				reader,
				timeService,
				fs,
				2,
			)

			err = releaseDir.Init(InitOptions{Sample: true})
			Expect(err).To(HaveOccurred())
			Expect(fs.FileExists("/dir/jobs/sample-job")).To(BeTrue())
		})

		It("regenerates the sample on retry after a packaging script failure", func() {
			gen := &packagingFailureGenerator{
				FSGenerator: NewFSGenerator("/dir", fs),
				fs:          fs,
				armed:       true,
			}

			releaseDir := NewFSReleaseDir(
				"/dir",
				config,
				gitRepo,
				blobsDir,
				gen,
				devReleases,
				finalReleases,
				finalIndicies,
				reader,
				timeService,
				fs,
				2,
			)

			err := releaseDir.Init(InitOptions{Sample: true})
			Expect(err).To(HaveOccurred())

			gen.armed = false

			err = releaseDir.Init(InitOptions{Sample: true, Force: true})
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/packages/sample-pkg/packaging")).To(Equal(
				"set -eu\n\ncp -a sample-pkg/. \"${BOSH_INSTALL_TARGET}/\"\n"))
			Expect(fs.ReadFileString("/dir/jobs/sample-job/templates/bin/sample-job")).To(ContainSubstring(
				"cat /var/vcap/packages/sample-pkg/file.txt"))
		})
	})

	Describe("GenerateJob", func() {
		It("delegates to generator", func() {
			gen.GenerateJobStub = func(opts GenerateJobOptions) error {
				Expect(opts).To(Equal(GenerateJobOptions{Name: "job1", Monit: true}))
				return errors.New("fake-err")
			}
			Expect(releaseDir.GenerateJob(GenerateJobOptions{Name: "job1", Monit: true})).To(Equal(errors.New("fake-err")))
		})
	})

	Describe("GeneratePackage", func() {
		It("delegates to generator", func() {
			gen.GeneratePackageStub = func(opts GeneratePackageOptions) error {
				Expect(opts).To(Equal(GeneratePackageOptions{Name: "pkg1", Files: []string{"pkg1/**/*"}}))
				return errors.New("fake-err")
			}
			Expect(releaseDir.GeneratePackage(GeneratePackageOptions{Name: "pkg1", Files: []string{"pkg1/**/*"}})).To(Equal(errors.New("fake-err")))
		})
	})

	Describe("Reset", func() {
		It("removes .blobs, blobs, .dev_builds and dev_releases", func() {
			err := fs.WriteFileString("/dir/.dev_builds/sub-dir", "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString("/dir/dev_releases/sub-dir", "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString("/dir/.blobs/sub-dir", "")
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString("/dir/blobs/sub-dir", "")
			Expect(err).ToNot(HaveOccurred())

			err = releaseDir.Reset()
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.FileExists("/dir/.dev_builds")).To(BeFalse())
			Expect(fs.FileExists("/dir/dev_releases")).To(BeFalse())
			Expect(fs.FileExists("/dir/.blobs")).To(BeFalse())
			Expect(fs.FileExists("/dir/blobs")).To(BeFalse())
		})

		It("returns error when deleting directory fails", func() {
			fs.RemoveAllStub = func(_ string) error { return errors.New("fake-err") }

			err := releaseDir.Reset()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})

	Describe("DefaultName", func() {
		It("delegates to config", func() {
			config.NameReturns("name", errors.New("fake-err"))

			name, err := releaseDir.DefaultName()
			Expect(name).To(Equal("name"))
			Expect(err).To(Equal(errors.New("fake-err")))
		})
	})

	Describe("NextFinalVersion", func() {
		It("returns incremented last final version for specific release name", func() {
			finalReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1")
				return &lastVer, nil
			}

			ver, err := releaseDir.NextFinalVersion("rel1")
			Expect(err).ToNot(HaveOccurred())
			Expect(ver.String()).To(Equal(semver.MustNewVersionFromString("1.2").String()))
		})

		It("returns '1' if there are no versions so that when it's finalized it will be '1'", func() {
			finalReleases.LastVersionReturns(nil, nil)

			ver, err := releaseDir.NextFinalVersion("rel1")
			Expect(err).ToNot(HaveOccurred())
			Expect(ver.String()).To(Equal(semver.MustNewVersionFromString("1").String()))
		})

		It("returns error if cannot find out last version", func() {
			finalReleases.LastVersionReturns(nil, errors.New("fake-err"))

			_, err := releaseDir.NextFinalVersion("rel1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if incrementing fails", func() {
			lastVer := semver.MustNewVersionFromString("a")
			finalReleases.LastVersionReturns(&lastVer, nil)

			_, err := releaseDir.NextFinalVersion("rel1")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Incrementing last final version"))
		})
	})

	Describe("NextDevVersion", func() {
		It("returns incremented last final version for specific release name", func() {
			finalReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1")
				return &lastVer, nil
			}

			ver, err := releaseDir.NextDevVersion("rel1", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ver.String()).To(Equal(semver.MustNewVersionFromString("1.1+dev.1").String()))
		})

		It("returns incremented last dev version for specific release name", func() {
			devReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1+dev.1")
				return &lastVer, nil
			}

			ver, err := releaseDir.NextDevVersion("rel1", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ver.String()).To(Equal(semver.MustNewVersionFromString("1.1+dev.2").String()))
		})

		It("returns timestamp-ed dev version for specific release name", func() {
			devReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1+dev.1")
				return &lastVer, nil
			}

			ver, err := releaseDir.NextDevVersion("rel1", true)
			Expect(err).ToNot(HaveOccurred())
			Expect(ver.String()).To(Equal(semver.MustNewVersionFromString("1.1+dev.1257894062").String()))
		})

		It("returns incremented greater dev version compared to final version for specific release name", func() {
			finalReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1")
				return &lastVer, nil
			}

			devReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1+dev.1")
				return &lastVer, nil
			}

			ver, err := releaseDir.NextDevVersion("rel1", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ver.String()).To(Equal(semver.MustNewVersionFromString("1.1+dev.2").String()))
		})

		It("returns incremented greater final version compared to dev version for specific release name", func() {
			finalReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.2")
				return &lastVer, nil
			}

			devReleases.LastVersionStub = func(name string) (*semver.Version, error) {
				Expect(name).To(Equal("rel1"))
				lastVer := semver.MustNewVersionFromString("1.1+dev.1")
				return &lastVer, nil
			}

			ver, err := releaseDir.NextDevVersion("rel1", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ver.String()).To(Equal(semver.MustNewVersionFromString("1.2+dev.1").String()))
		})

		It("returns '0+dev.1' if there are no dev or final versions", func() {
			ver, err := releaseDir.NextDevVersion("rel1", false)
			Expect(err).ToNot(HaveOccurred())
			Expect(ver.String()).To(Equal(semver.MustNewVersionFromString("0+dev.1").String()))
		})

		It("returns first timestamp-ed dev version if there are no dev or final versions", func() {
			ver, err := releaseDir.NextDevVersion("rel1", true)
			Expect(err).ToNot(HaveOccurred())
			Expect(ver.String()).To(Equal(semver.MustNewVersionFromString("0+dev.1257894062").String()))
		})

		It("returns error if cannot find out last dev version", func() {
			devReleases.LastVersionReturns(nil, errors.New("fake-err"))

			_, err := releaseDir.NextDevVersion("rel1", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if cannot find out last final version", func() {
			finalReleases.LastVersionReturns(nil, errors.New("fake-err"))

			_, err := releaseDir.NextDevVersion("rel1", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if incrementing fails", func() {
			lastVer := semver.MustNewVersionFromString("1+a")
			finalReleases.LastVersionReturns(&lastVer, nil)

			_, err := releaseDir.NextDevVersion("rel1", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Incrementing last dev version"))
		})
	})

	Describe("FindRelease", func() {
		var (
			expectedRelease *fakerel.FakeRelease
		)

		BeforeEach(func() {
			expectedRelease = &fakerel.FakeRelease{
				NameStub:          func() string { return "rel1" },
				NoCompressionStub: func() bool { return false },
			}
		})

		Context("when name and version are not specified", func() {
			BeforeEach(func() {
				config.NameReturns("rel1", nil)
			})

			It("returns last final release for specific release name", func() {
				finalReleases.LastVersionStub = func(name string) (*semver.Version, error) {
					Expect(name).To(Equal("rel1"))
					lastVer := semver.MustNewVersionFromString("1.1")
					return &lastVer, nil
				}

				finalReleases.ManifestPathStub = func(name, ver string) string {
					Expect(name).To(Equal("rel1"))
					Expect(ver).To(Equal("1.1"))
					return "manifest-path"
				}

				reader.ReadStub = func(path string) (boshrel.Release, error) {
					Expect(path).To(Equal("manifest-path"))
					return expectedRelease, nil
				}

				release, err := releaseDir.FindRelease("", semver.Version{})
				Expect(err).ToNot(HaveOccurred())
				Expect(release).To(Equal(expectedRelease))
			})

			It("returns last dev release for specific release name", func() {
				devReleases.LastVersionStub = func(name string) (*semver.Version, error) {
					Expect(name).To(Equal("rel1"))
					lastVer := semver.MustNewVersionFromString("1.1+dev.1")
					return &lastVer, nil
				}

				devReleases.ManifestPathStub = func(name, ver string) string {
					Expect(name).To(Equal("rel1"))
					Expect(ver).To(Equal("1.1+dev.1"))
					return "manifest-path"
				}

				reader.ReadStub = func(path string) (boshrel.Release, error) {
					Expect(path).To(Equal("manifest-path"))
					return expectedRelease, nil
				}

				release, err := releaseDir.FindRelease("", semver.Version{})
				Expect(err).ToNot(HaveOccurred())
				Expect(release).To(Equal(expectedRelease))
			})

			It("returns greater dev release compared to final release for specific release name", func() {
				finalReleases.LastVersionStub = func(name string) (*semver.Version, error) {
					Expect(name).To(Equal("rel1"))
					lastVer := semver.MustNewVersionFromString("1.1")
					return &lastVer, nil
				}

				devReleases.LastVersionStub = func(name string) (*semver.Version, error) {
					Expect(name).To(Equal("rel1"))
					lastVer := semver.MustNewVersionFromString("1.1+dev.1")
					return &lastVer, nil
				}

				devReleases.ManifestPathStub = func(name, ver string) string {
					Expect(name).To(Equal("rel1"))
					Expect(ver).To(Equal("1.1+dev.1"))
					return "manifest-path"
				}

				reader.ReadStub = func(path string) (boshrel.Release, error) {
					Expect(path).To(Equal("manifest-path"))
					return expectedRelease, nil
				}

				release, err := releaseDir.FindRelease("", semver.Version{})
				Expect(err).ToNot(HaveOccurred())
				Expect(release).To(Equal(expectedRelease))
			})

			It("returns greater final release compared to dev release for specific release name", func() {
				finalReleases.LastVersionStub = func(name string) (*semver.Version, error) {
					Expect(name).To(Equal("rel1"))
					lastVer := semver.MustNewVersionFromString("1.2")
					return &lastVer, nil
				}

				devReleases.LastVersionStub = func(name string) (*semver.Version, error) {
					Expect(name).To(Equal("rel1"))
					lastVer := semver.MustNewVersionFromString("1.1+dev.1")
					return &lastVer, nil
				}

				finalReleases.ManifestPathStub = func(name, ver string) string {
					Expect(name).To(Equal("rel1"))
					Expect(ver).To(Equal("1.2"))
					return "manifest-path"
				}

				reader.ReadStub = func(path string) (boshrel.Release, error) {
					Expect(path).To(Equal("manifest-path"))
					return expectedRelease, nil
				}

				release, err := releaseDir.FindRelease("", semver.Version{})
				Expect(err).ToNot(HaveOccurred())
				Expect(release).To(Equal(expectedRelease))
			})

			It("returns error if there are no dev or final versions", func() {
				_, err := releaseDir.FindRelease("", semver.Version{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Expected to find at least one dev or final version"))
			})

			It("returns error if cannot find out last dev version", func() {
				devReleases.LastVersionReturns(nil, errors.New("fake-err"))

				_, err := releaseDir.FindRelease("", semver.Version{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})

			It("returns error if cannot find out last final version", func() {
				finalReleases.LastVersionReturns(nil, errors.New("fake-err"))

				_, err := releaseDir.FindRelease("", semver.Version{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})

			It("retuns error if cannot determine final name", func() {
				config.NameReturns("", errors.New("fake-err"))

				_, err := releaseDir.FindRelease("", semver.Version{})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("fake-err"))
			})
		})

		Context("when name and version is specified", func() {
			It("returns final release for specific release name and version", func() {
				finalReleases.ManifestPathStub = func(name, ver string) string {
					Expect(name).To(Equal("rel1"))
					Expect(ver).To(Equal("1.1"))
					return "manifest-path"
				}

				reader.ReadStub = func(path string) (boshrel.Release, error) {
					Expect(path).To(Equal("manifest-path"))
					return expectedRelease, nil
				}

				release, err := releaseDir.FindRelease("rel1", semver.MustNewVersionFromString("1.1"))
				Expect(err).ToNot(HaveOccurred())
				Expect(release).To(Equal(expectedRelease))
			})
		})
	})

	Describe("BuildRelease", func() {
		var (
			ver             semver.Version
			expectedRelease *fakerel.FakeRelease
		)

		BeforeEach(func() {
			ver = semver.MustNewVersionFromString("1.1")

			expectedRelease = &fakerel.FakeRelease{
				NameStub: func() string { return "rel1" },
				ManifestStub: func() boshman.Manifest {
					return boshman.Manifest{Name: "rel1"}
				},
				NoCompressionStub: func() bool { return false },
			}
		})

		It("builds release", func() {
			var ops []string

			gitRepo.MustNotBeDirtyStub = func(force bool) (bool, error) {
				ops = append(ops, "dirty")
				return true, nil
			}

			gitRepo.LastCommitSHAReturns("commit", nil)

			blobsDir.SyncBlobsStub = func(numOfParallelWorkers int) error {
				ops = append(ops, "blobs")
				return nil
			}

			reader.ReadStub = func(path string) (boshrel.Release, error) {
				Expect(path).To(Equal("/dir"))
				ops = append(ops, "read")
				return expectedRelease, nil
			}

			devReleases.AddStub = func(manifest boshman.Manifest) error {
				Expect(manifest).To(Equal(boshman.Manifest{Name: "rel1"}))
				ops = append(ops, "manifest")
				return nil
			}

			release, err := releaseDir.BuildRelease("rel1", ver, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(release).To(Equal(expectedRelease))

			Expect(expectedRelease.SetNameArgsForCall(0)).To(Equal("rel1"))
			Expect(expectedRelease.SetVersionArgsForCall(0)).To(Equal("1.1"))
			Expect(expectedRelease.SetCommitHashArgsForCall(0)).To(Equal("commit"))
			Expect(expectedRelease.SetUncommittedChangesArgsForCall(0)).To(BeTrue())

			Expect(ops).To(Equal([]string{"dirty", "blobs", "read", "manifest"}))
		})

		It("returns error if git is dirty and force is not set", func() {
			gitRepo.MustNotBeDirtyReturns(true, errors.New("dirty"))

			_, err := releaseDir.BuildRelease("rel1", ver, false)
			Expect(err).To(Equal(errors.New("dirty")))
		})

		It("returns error if last commit cannot be retrieved", func() {
			gitRepo.LastCommitSHAReturns("", errors.New("fake-err"))

			_, err := releaseDir.BuildRelease("rel1", ver, false)
			Expect(err).To(Equal(errors.New("fake-err")))
		})

		It("returns error if reading release", func() {
			reader.ReadReturns(nil, errors.New("fake-err"))

			_, err := releaseDir.BuildRelease("rel1", ver, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if adding dev release fails", func() {
			reader.ReadReturns(expectedRelease, nil)
			devReleases.AddReturns(errors.New("fake-err"))

			_, err := releaseDir.BuildRelease("rel1", ver, false)
			Expect(err).To(Equal(errors.New("fake-err")))
		})

		It("sets no_compression to false in manifest when noCompression is false", func() {
			gitRepo.MustNotBeDirtyReturns(true, nil)
			gitRepo.LastCommitSHAReturns("commit", nil)
			blobsDir.SyncBlobsReturns(nil)
			reader.ReadReturns(expectedRelease, nil)

			var capturedManifest boshman.Manifest
			devReleases.AddStub = func(manifest boshman.Manifest) error {
				capturedManifest = manifest
				return nil
			}

			_, err := releaseDir.BuildRelease("rel1", ver, false)
			Expect(err).ToNot(HaveOccurred())

			Expect(capturedManifest.NoCompression).To(BeFalse())
			Expect(expectedRelease.SetNoCompressionArgsForCall(0)).To(BeFalse())
		})

		It("sets no_compression to true in manifest when noCompression is true", func() {
			config.NoCompressionReturns(true)
			expectedRelease.ManifestStub = func() boshman.Manifest {
				return boshman.Manifest{Name: "rel1", NoCompression: true}
			}
			gitRepo.MustNotBeDirtyReturns(true, nil)
			gitRepo.LastCommitSHAReturns("commit", nil)
			blobsDir.SyncBlobsReturns(nil)
			reader.ReadReturns(expectedRelease, nil)

			var capturedManifest boshman.Manifest
			devReleases.AddStub = func(manifest boshman.Manifest) error {
				capturedManifest = manifest
				return nil
			}

			_, err := releaseDir.BuildRelease("rel1", ver, false)
			Expect(err).ToNot(HaveOccurred())

			Expect(capturedManifest.NoCompression).To(BeTrue())
			Expect(expectedRelease.SetNoCompressionArgsForCall(0)).To(BeTrue())
		})
	})

	Describe("VendorPackage", func() {
		It("finalizes given package and its dependencies and records vendored package manifest locks", func() {
			pkg1Res := &fakeres.FakeResource{
				NameStub:        func() string { return "pkg1-name" },
				FingerprintStub: func() string { return "pkg1-fp" },
			}
			pkg2Res := &fakeres.FakeResource{
				NameStub:        func() string { return "pkg2-name" },
				FingerprintStub: func() string { return "pkg2-fp" },
			}
			pkg3Res := &fakeres.FakeResource{
				NameStub:        func() string { return "pkg3-name" },
				FingerprintStub: func() string { return "pkg3-fp" },
			}
			pkg4Res := &fakeres.FakeResource{
				NameStub:        func() string { return "pkg4-name" },
				FingerprintStub: func() string { return "pkg4-fp" },
			}

			pkg1 := boshpkg.NewPackage(pkg1Res, []string{"pkg2-name", "pkg3-name"})
			pkg2 := boshpkg.NewPackage(pkg2Res, []string{"pkg4-name"})
			pkg3 := boshpkg.NewPackage(pkg3Res, []string{"pkg4-name"})
			pkg4 := boshpkg.NewPackage(pkg4Res, nil)
			err := pkg1.AttachDependencies([]*boshpkg.Package{pkg2, pkg3})
			Expect(err).ToNot(HaveOccurred())
			err = pkg2.AttachDependencies([]*boshpkg.Package{pkg4})
			Expect(err).ToNot(HaveOccurred())
			err = pkg3.AttachDependencies([]*boshpkg.Package{pkg4})
			Expect(err).ToNot(HaveOccurred())

			// previous content will be overwritten
			Expect(fs.WriteFileString("/dir/packages/pkg1-name/spec", "old-spec")).ToNot(HaveOccurred())
			Expect(fs.WriteFileString("/dir/packages/pkg1-name/packaging", "old-packaging")).ToNot(HaveOccurred())
			Expect(fs.WriteFileString("/dir/packages/pkg2-name/spec.lock", "old-spec-lock")).ToNot(HaveOccurred())

			err = releaseDir.VendorPackage(pkg1, "")
			Expect(err).ToNot(HaveOccurred())

			// recorded files
			Expect(fs.FileExists("/dir/packages/pkg1-name/spec")).To(BeFalse())
			Expect(fs.FileExists("/dir/packages/pkg1-name/packaging")).To(BeFalse())
			Expect(fs.ReadFileString("/dir/packages/pkg1-name/spec.lock")).To(Equal(`name: pkg1-name
fingerprint: pkg1-fp
dependencies:
- pkg2-name
- pkg3-name
`))

			Expect(fs.ReadFileString("/dir/packages/pkg2-name/spec.lock")).To(Equal(`name: pkg2-name
fingerprint: pkg2-fp
dependencies:
- pkg4-name
`))

			Expect(fs.ReadFileString("/dir/packages/pkg3-name/spec.lock")).To(Equal(`name: pkg3-name
fingerprint: pkg3-fp
dependencies:
- pkg4-name
`))

			Expect(fs.ReadFileString("/dir/packages/pkg4-name/spec.lock")).To(Equal(`name: pkg4-name
fingerprint: pkg4-fp
`))

			// Use == for pointer equality
			Expect(finalIndicies.Packages).ToNot(BeNil())

			Expect(pkg1Res.FinalizeCallCount()).To(Equal(1))
			Expect(pkg1Res.FinalizeArgsForCall(0) == finalIndicies.Packages).To(BeTrue())

			Expect(pkg2Res.FinalizeCallCount()).To(Equal(1))
			Expect(pkg2Res.FinalizeArgsForCall(0) == finalIndicies.Packages).To(BeTrue())

			Expect(pkg3Res.FinalizeCallCount()).To(Equal(1))
			Expect(pkg3Res.FinalizeArgsForCall(0) == finalIndicies.Packages).To(BeTrue())

			Expect(pkg4Res.FinalizeCallCount()).To(Equal(1))
			Expect(pkg4Res.FinalizeArgsForCall(0) == finalIndicies.Packages).To(BeTrue())
		})

		It("reliably vendors packages with prefix inlcuding their dependencies", func() {
			_, err := fs.ReadFileString("/dir/packages/prefixed-pkg1-name/spec.lock")

			Expect(err).To(HaveOccurred())
			pkg2Res := resource.NewResourceWithBuiltArchive("pkg2-name", "pkg2-fp", "/test/something", "something")
			pkg3Res := resource.NewResourceWithBuiltArchive("pkg3-name", "pkg3-fp", "/test/something", "something")
			pkg2 := boshpkg.NewPackage(pkg2Res, []string{})
			pkg3 := boshpkg.NewPackage(pkg3Res, []string{})

			pkg1Res := resource.NewResourceWithBuiltArchive("pkg1-name", "pkg1-fp", "/test/something", "something")
			pkg1 := boshpkg.NewPackage(pkg1Res, []string{"pkg2-name", "pkg3-name"})

			err = pkg1.AttachDependencies([]*boshpkg.Package{pkg2, pkg3})
			Expect(err).ToNot(HaveOccurred())

			err = releaseDir.VendorPackage(pkg1, "prefixed")
			Expect(err).ToNot(HaveOccurred())

			Expect(fs.ReadFileString("/dir/packages/prefixed-pkg1-name/spec.lock")).To(Equal(`name: prefixed-pkg1-name
fingerprint: pkg1-fp
dependencies:
- prefixed-pkg2-name
- prefixed-pkg3-name
`))
		})

		It("finalize given package with prefix", func() {
			pkg1Res := resource.NewResourceWithBuiltArchive("pkg1-name", "pkg1-fp", "/test/something", "something")
			pkg1 := boshpkg.NewPackage(pkg1Res, nil)

			err := releaseDir.VendorPackage(pkg1, "prefix")
			Expect(err).ToNot(HaveOccurred())
			Expect(fs.ReadFileString("/dir/packages/prefix-pkg1-name/spec.lock")).To(Equal(`name: prefix-pkg1-name
fingerprint: pkg1-fp
`))
		})

		It("returns error if package finalize fails", func() {
			pkg1Res := &fakeres.FakeResource{
				NameStub:        func() string { return "pkg1-name" },
				FingerprintStub: func() string { return "pkg1-fp" },
			}
			pkg1 := boshpkg.NewPackage(pkg1Res, nil)

			pkg1Res.FinalizeReturns(errors.New("fake-err"))

			err := releaseDir.VendorPackage(pkg1, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if removing old package dir fails", func() {
			pkg1Res := &fakeres.FakeResource{
				NameStub:        func() string { return "pkg1-name" },
				FingerprintStub: func() string { return "pkg1-fp" },
			}
			pkg1 := boshpkg.NewPackage(pkg1Res, nil)

			fs.RemoveAllStub = func(path string) error { return errors.New("fake-err") }

			err := releaseDir.VendorPackage(pkg1, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})

		It("returns error if serializing manifest lock fails", func() {
			pkg1Res := &fakeres.FakeResource{
				NameStub: func() string { return "pkg1-name" },
			}
			pkg1 := boshpkg.NewPackage(pkg1Res, nil)

			err := releaseDir.VendorPackage(pkg1, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Marshaling vendored package 'pkg1-name' spec lock"))
		})

		It("returns error if writing manifest lock fails", func() {
			pkg1Res := &fakeres.FakeResource{
				NameStub:        func() string { return "pkg1-name" },
				FingerprintStub: func() string { return "pkg1-fp" },
			}
			pkg1 := boshpkg.NewPackage(pkg1Res, nil)

			fs.WriteFileErrors["/dir/packages/pkg1-name/spec.lock"] = errors.New("fake-err")

			err := releaseDir.VendorPackage(pkg1, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-err"))
		})
	})

	Describe("FinalizeRelease", func() {
		var (
			release *fakerel.FakeRelease
		)

		BeforeEach(func() {
			release = &fakerel.FakeRelease{
				NameStub:    func() string { return "rel1" },
				VersionStub: func() string { return "ver1" },
				ManifestStub: func() boshman.Manifest {
					return boshman.Manifest{Name: "rel1"}
				},
				NoCompressionStub: func() bool { return false },
			}
		})

		It("finalizes release", func() {
			var ops []string

			gitRepo.MustNotBeDirtyStub = func(force bool) (bool, error) {
				ops = append(ops, "dirty")
				return true, nil
			}

			finalReleases.ContainsStub = func(rel boshrel.Release) (bool, error) {
				Expect(rel).To(Equal(release))
				ops = append(ops, "check")
				return false, nil
			}

			release.FinalizeStub = func(indicies boshrel.ArchiveIndicies, parallel int) error {
				Expect(parallel).To(Equal(2))
				Expect(indicies.Jobs).To(Equal(finalIndicies.Jobs)) // unique check
				ops = append(ops, "finalize")
				return nil
			}

			finalReleases.AddStub = func(manifest boshman.Manifest) error {
				Expect(manifest).To(Equal(boshman.Manifest{Name: "rel1"}))
				ops = append(ops, "manifest")
				return nil
			}

			err := releaseDir.FinalizeRelease(release, false)
			Expect(err).ToNot(HaveOccurred())

			Expect(ops).To(Equal([]string{"dirty", "check", "finalize", "manifest"}))
		})

		It("returns error if git is dirty and force is not set", func() {
			gitRepo.MustNotBeDirtyReturns(true, errors.New("dirty"))

			err := releaseDir.FinalizeRelease(release, false)
			Expect(err).To(Equal(errors.New("dirty")))
		})

		It("returns error if checking for a final release fails", func() {
			finalReleases.ContainsReturns(false, errors.New("fake-err"))

			err := releaseDir.FinalizeRelease(release, false)
			Expect(err).To(Equal(errors.New("fake-err")))
		})

		It("returns error if final release index already contains this name/ver", func() {
			finalReleases.ContainsReturns(true, nil)

			err := releaseDir.FinalizeRelease(release, false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("Release 'rel1' version 'ver1' already exists"))
		})

		It("returns error if adding final release fails", func() {
			finalReleases.AddReturns(errors.New("fake-err"))

			err := releaseDir.FinalizeRelease(release, false)
			Expect(err).To(Equal(errors.New("fake-err")))
		})
	})

	Describe("CheckNoCompressionMismatch", func() {
		It("returns no mismatch when release has no_compression: true", func() {
			config.NoCompressionReturns(true)

			hasMismatch, packages, err := releaseDir.CheckNoCompressionMismatch()
			Expect(err).ToNot(HaveOccurred())
			Expect(hasMismatch).To(BeFalse())
			Expect(packages).To(BeEmpty())
		})

		It("returns no mismatch when no packages have no_compression: true", func() {
			config.NoCompressionReturns(false)

			err := fs.MkdirAll("/dir/packages/pkg1", os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString("/dir/packages/pkg1/spec", `---
name: pkg1
dependencies: []
files: []
`)
			Expect(err).ToNot(HaveOccurred())

			fs.SetGlob("/dir/packages/*", []string{"/dir/packages/pkg1"})

			hasMismatch, packages, err := releaseDir.CheckNoCompressionMismatch()
			Expect(err).ToNot(HaveOccurred())
			Expect(hasMismatch).To(BeFalse())
			Expect(packages).To(BeEmpty())
		})

		It("returns no mismatch when packages have no_compression: true and final.yml also has it", func() {
			config.NoCompressionReturns(true)

			err := fs.MkdirAll("/dir/packages/pkg1", os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString("/dir/packages/pkg1/spec", `---
name: pkg1
dependencies: []
files: []
no_compression: true
`)
			Expect(err).ToNot(HaveOccurred())

			fs.SetGlob("/dir/packages/*", []string{"/dir/packages/pkg1"})

			hasMismatch, packages, err := releaseDir.CheckNoCompressionMismatch()
			Expect(err).ToNot(HaveOccurred())
			Expect(hasMismatch).To(BeFalse())
			Expect(packages).To(BeEmpty())
		})

		It("returns mismatch when packages have no_compression: true but final.yml doesn't", func() {
			config.NoCompressionReturns(false)

			err := fs.MkdirAll("/dir/packages/pkg1", os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString("/dir/packages/pkg1/spec", `---
name: pkg1
dependencies: []
files: []
no_compression: true
`)
			Expect(err).ToNot(HaveOccurred())

			fs.SetGlob("/dir/packages/*", []string{"/dir/packages/pkg1"})

			hasMismatch, packages, err := releaseDir.CheckNoCompressionMismatch()
			Expect(err).ToNot(HaveOccurred())
			Expect(hasMismatch).To(BeTrue())
			Expect(packages).To(Equal([]string{"pkg1"}))
		})

		It("returns mismatch with multiple packages having no_compression: true", func() {
			config.NoCompressionReturns(false)

			err := fs.MkdirAll("/dir/packages/pkg1", os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString("/dir/packages/pkg1/spec", `---
name: pkg1
dependencies: []
files: []
no_compression: true
`)
			Expect(err).ToNot(HaveOccurred())

			err = fs.MkdirAll("/dir/packages/pkg2", os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString("/dir/packages/pkg2/spec", `---
name: pkg2
dependencies: []
files: []
no_compression: true
`)
			Expect(err).ToNot(HaveOccurred())

			err = fs.MkdirAll("/dir/packages/pkg3", os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString("/dir/packages/pkg3/spec", `---
name: pkg3
dependencies: []
files: []
`)
			Expect(err).ToNot(HaveOccurred())

			fs.SetGlob("/dir/packages/*", []string{"/dir/packages/pkg1", "/dir/packages/pkg2", "/dir/packages/pkg3"})

			hasMismatch, packages, err := releaseDir.CheckNoCompressionMismatch()
			Expect(err).ToNot(HaveOccurred())
			Expect(hasMismatch).To(BeTrue())
			Expect(packages).To(ConsistOf("pkg1", "pkg2"))
		})

		It("skips packages without spec files", func() {
			config.NoCompressionReturns(false)

			err := fs.MkdirAll("/dir/packages/pkg1", os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			// No spec file

			fs.SetGlob("/dir/packages/*", []string{"/dir/packages/pkg1"})

			hasMismatch, packages, err := releaseDir.CheckNoCompressionMismatch()
			Expect(err).ToNot(HaveOccurred())
			Expect(hasMismatch).To(BeFalse())
			Expect(packages).To(BeEmpty())
		})

		It("skips invalid spec files and continues checking others", func() {
			config.NoCompressionReturns(false)

			err := fs.MkdirAll("/dir/packages/pkg1", os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString("/dir/packages/pkg1/spec", `invalid yaml: [`)
			Expect(err).ToNot(HaveOccurred())

			err = fs.MkdirAll("/dir/packages/pkg2", os.ModePerm)
			Expect(err).ToNot(HaveOccurred())
			err = fs.WriteFileString("/dir/packages/pkg2/spec", `---
name: pkg2
dependencies: []
files: []
no_compression: true
`)
			Expect(err).ToNot(HaveOccurred())

			fs.SetGlob("/dir/packages/*", []string{"/dir/packages/pkg1", "/dir/packages/pkg2"})

			hasMismatch, packages, err := releaseDir.CheckNoCompressionMismatch()
			Expect(err).ToNot(HaveOccurred())
			Expect(hasMismatch).To(BeTrue())
			Expect(packages).To(Equal([]string{"pkg2"}))
		})

		It("returns error if glob fails", func() {
			config.NoCompressionReturns(false)

			fs.GlobErr = errors.New("glob-error")

			hasMismatch, packages, err := releaseDir.CheckNoCompressionMismatch()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("glob-error"))
			Expect(hasMismatch).To(BeFalse())
			Expect(packages).To(BeNil())
		})
	})
})
