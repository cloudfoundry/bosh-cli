package release

import (
	"os"
	gopath "path"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshcmd "github.com/cloudfoundry/bosh-utils/fileutil"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"

	boshjob "github.com/cloudfoundry/bosh-init/release/job"
	boshlic "github.com/cloudfoundry/bosh-init/release/license"
	boshpkg "github.com/cloudfoundry/bosh-init/release/pkg"
)

type ArchiveWriter struct {
	compressor boshcmd.Compressor
	fs         boshsys.FileSystem

	logTag string
	logger boshlog.Logger
}

func NewArchiveWriter(compressor boshcmd.Compressor, fs boshsys.FileSystem, logger boshlog.Logger) ArchiveWriter {
	return ArchiveWriter{compressor: compressor, fs: fs, logTag: "release.ArchiveWriter", logger: logger}
}

func (w ArchiveWriter) Write(release Release, pkgFpsToSkip []string) (string, error) {
	stagingDir, err := w.fs.TempDir("bosh-release")
	if err != nil {
		return "", bosherr.WrapErrorf(err, "Creating staging release dir")
	}

	w.logger.Info(w.logTag, "Writing release tarball into '%s'", stagingDir)

	manifestBytes, err := yaml.Marshal(release.Manifest())
	if err != nil {
		w.cleanUp(stagingDir)
		return "", bosherr.WrapError(err, "Marshalling release manifest")
	}

	manifestPath := gopath.Join(stagingDir, "release.MF")

	err = w.fs.WriteFile(manifestPath, manifestBytes)
	if err != nil {
		w.cleanUp(stagingDir)
		return "", bosherr.WrapErrorf(err, "Writing release manifest '%s'", manifestPath)
	}

	err = w.writePackages(release.Packages(), pkgFpsToSkip, stagingDir)
	if err != nil {
		w.cleanUp(stagingDir)
		return "", bosherr.WrapError(err, "Writing packages")
	}

	err = w.writeCompiledPackages(release.CompiledPackages(), pkgFpsToSkip, stagingDir)
	if err != nil {
		w.cleanUp(stagingDir)
		return "", bosherr.WrapError(err, "Writing compiled packages")
	}

	err = w.writeJobs(release.Jobs(), stagingDir)
	if err != nil {
		w.cleanUp(stagingDir)
		return "", bosherr.WrapError(err, "Writing jobs")
	}

	err = w.writeLicense(release.License(), stagingDir)
	if err != nil {
		w.cleanUp(stagingDir)
		return "", bosherr.WrapError(err, "Writing license")
	}

	path, err := w.compressor.CompressFilesInDir(stagingDir)
	if err != nil {
		w.cleanUp(stagingDir)
		return "", bosherr.WrapError(err, "Compressing release")
	}

	w.cleanUp(stagingDir)

	return path, nil
}

func (w ArchiveWriter) cleanUp(stagingDir string) {
	removeErr := w.fs.RemoveAll(stagingDir)
	if removeErr != nil {
		w.logger.Error(w.logTag, "Failed to remove staging dir for release: %s", removeErr.Error())
	}
}

func (w ArchiveWriter) writeJobs(jobs []*boshjob.Job, stagingDir string) error {
	if len(jobs) == 0 {
		return nil
	}

	jobsPath := gopath.Join(stagingDir, "jobs")

	err := w.fs.MkdirAll(jobsPath, os.ModePerm)
	if err != nil {
		return bosherr.WrapError(err, "Creating jobs/")
	}

	for _, job := range jobs {
		err := w.fs.CopyFile(job.ArchivePath(), gopath.Join(jobsPath, job.Name()+".tgz"))
		if err != nil {
			return bosherr.WrapErrorf(err, "Copying job '%s' archive into staging dir", job.Name())
		}
	}

	return nil
}

func (w ArchiveWriter) writePackages(packages []*boshpkg.Package, pkgFpsToSkip []string, stagingDir string) error {
	if len(packages) == 0 {
		return nil
	}

	pkgsPath := gopath.Join(stagingDir, "packages")

	err := w.fs.MkdirAll(pkgsPath, os.ModePerm)
	if err != nil {
		return bosherr.WrapError(err, "Creating packages/")
	}

	for _, pkg := range packages {
		if w.shouldSkip(pkg.Fingerprint(), pkgFpsToSkip) {
			w.logger.Debug(w.logTag, "Package '%s' was filtered out", pkg.Name())
		} else {
			err := w.fs.CopyFile(pkg.ArchivePath(), gopath.Join(pkgsPath, pkg.Name()+".tgz"))
			if err != nil {
				return bosherr.WrapErrorf(err, "Copying package '%s' archive into staging dir", pkg.Name())
			}
		}
	}

	return nil
}

func (w ArchiveWriter) writeCompiledPackages(compiledPkgs []*boshpkg.CompiledPackage, pkgFpsToSkip []string, stagingDir string) error {
	if len(compiledPkgs) == 0 {
		return nil
	}

	pkgsPath := gopath.Join(stagingDir, "compiled_packages")

	err := w.fs.MkdirAll(pkgsPath, os.ModePerm)
	if err != nil {
		return bosherr.WrapError(err, "Creating compiled_packages/")
	}

	for _, compiledPkg := range compiledPkgs {
		if w.shouldSkip(compiledPkg.Fingerprint(), pkgFpsToSkip) {
			w.logger.Debug(w.logTag, "Compiled package '%s' was filtered out", compiledPkg.Name())
		} else {
			err := w.fs.CopyFile(compiledPkg.ArchivePath(), gopath.Join(pkgsPath, compiledPkg.Name()+".tgz"))
			if err != nil {
				return bosherr.WrapErrorf(err, "Copying compiled package '%s' archive into staging dir", compiledPkg.Name())
			}
		}
	}

	return nil
}

func (w ArchiveWriter) writeLicense(license *boshlic.License, stagingDir string) error {
	if license == nil {
		return nil
	}

	err := w.fs.CopyFile(license.ArchivePath(), gopath.Join(stagingDir, "license.tgz"))
	if err != nil {
		return bosherr.WrapError(err, "Copying license archive into staging dir")
	}

	err = w.compressor.DecompressFileToDir(license.ArchivePath(), stagingDir, boshcmd.CompressorOptions{})
	if err != nil {
		return bosherr.WrapErrorf(err, "Decompressing license archive into staging dir")
	}

	return nil
}

func (w ArchiveWriter) shouldSkip(fp string, pkgFpsToSkip []string) bool {
	for _, pkgFp := range pkgFpsToSkip {
		if fp == pkgFp {
			return true
		}
	}
	return false
}
