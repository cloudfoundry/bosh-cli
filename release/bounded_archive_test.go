package release_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	util "github.com/cloudfoundry/bosh-cli/v7/common/util"
	release "github.com/cloudfoundry/bosh-cli/v7/release"
	job "github.com/cloudfoundry/bosh-cli/v7/release/job"
	lic "github.com/cloudfoundry/bosh-cli/v7/release/license"
	manifest "github.com/cloudfoundry/bosh-cli/v7/release/manifest"
	pkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	fakerel "github.com/cloudfoundry/bosh-cli/v7/release/releasefakes"
	resource "github.com/cloudfoundry/bosh-cli/v7/release/resource"
	fileutil "github.com/cloudfoundry/bosh-utils/fileutil"
	logger "github.com/cloudfoundry/bosh-utils/logger"
	system "github.com/cloudfoundry/bosh-utils/system"
)

func fixtureArchive(t *testing.T, headers []*tar.Header, payloads []string, compressed bool) string {
	t.Helper()
	var raw bytes.Buffer
	writer := tar.NewWriter(&raw)
	for i, header := range headers {
		if err := writer.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if header.Typeflag == tar.TypeReg {
			if _, err := writer.Write([]byte(payloads[i])); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if compressed {
		var zipped bytes.Buffer
		zip := gzip.NewWriter(&zipped)
		if _, err := zip.Write(raw.Bytes()); err != nil {
			t.Fatal(err)
		}
		if err := zip.Close(); err != nil {
			t.Fatal(err)
		}
		raw = zipped
	}
	filename := filepath.Join(t.TempDir(), "archive.tgz")
	if err := os.WriteFile(filename, raw.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return filename
}
func boundedCompressor() fileutil.Compressor {
	return fileutil.NewTarballCompressor(nil, system.NewOsFileSystem(logger.NewLogger(logger.LevelNone)))
}

func TestReaderAndLicenseWriterUseConfiguredTarLimit(t *testing.T) {
	t.Setenv("BOSH_RELEASE_MAX_EXPANDED_BYTES", "16")
	t.Setenv("BOSH_RELEASE_MAX_ARCHIVE_MEMBERS", "10")
	archive := fixtureArchive(t, []*tar.Header{{Name: "oversized", Mode: 0600, Size: 17, Typeflag: tar.TypeReg}}, []string{strings.Repeat("x", 17)}, true)
	log := logger.NewLogger(logger.LevelNone)
	fs := system.NewOsFileSystem(log)
	reader := release.NewArchiveReader(nil, nil, boundedCompressor(), fs, log)
	if _, err := reader.Read(archive); err == nil || !strings.Contains(err.Error(), "expanded bytes") {
		t.Fatal("release reader lacks limit", err)
	}
	item := &fakerel.FakeRelease{}
	item.LicenseReturns(lic.NewLicense(resource.NewResourceWithBuiltArchive("license", "fp", archive, "")))
	writer := release.NewArchiveWriter(boundedCompressor(), fs, log)
	if _, err := writer.Write(item, nil); err == nil || !strings.Contains(err.Error(), "expanded bytes") {
		t.Fatal("license writer lacks limit", err)
	}
}
func TestNestedJobAndPackageReadersUseLimits(t *testing.T) {
	t.Setenv("BOSH_RELEASE_MAX_EXPANDED_BYTES", "18")
	t.Setenv("BOSH_RELEASE_MAX_ARCHIVE_MEMBERS", "11")
	archive := fixtureArchive(t, []*tar.Header{{Name: "oversized", Size: 19, Mode: 0600, Typeflag: tar.TypeReg}}, []string{strings.Repeat("x", 19)}, true)
	fs := system.NewOsFileSystem(logger.NewLogger(logger.LevelNone))
	if _, err := job.NewArchiveReaderImpl(true, boundedCompressor(), fs).Read(manifest.JobRef{Name: "job"}, archive); err == nil || !strings.Contains(err.Error(), "expanded bytes") {
		t.Fatal("nested job bypassed limits", err)
	}
	if _, err := pkg.NewArchiveReaderImpl(true, boundedCompressor(), fs).Read(manifest.PackageRef{Name: "pkg"}, archive); err == nil || !strings.Contains(err.Error(), "expanded bytes") {
		t.Fatal("nested package bypassed limits", err)
	}
}
func TestRelatedArchivesShareCumulativeBudget(t *testing.T) {
	t.Setenv("BOSH_RELEASE_MAX_EXPANDED_BYTES", "9")
	t.Setenv("BOSH_RELEASE_MAX_ARCHIVE_MEMBERS", "7")
	for i, size := range []int64{8, 2} {
		archive := fixtureArchive(t, []*tar.Header{{Name: "data", Size: size, Mode: 0600, Typeflag: tar.TypeReg}}, []string{strings.Repeat("x", int(size))}, true)
		destination := t.TempDir()
		options, err := util.ArchiveExtractionOptions(destination)
		if err != nil {
			t.Fatal(err)
		}
		err = boundedCompressor().DecompressFileToDir(archive, destination, options)
		if i == 0 && err != nil {
			t.Fatal(err)
		}
		if i == 1 {
			if err == nil || !strings.Contains(err.Error(), "combined") {
				t.Fatal("nested extraction reset shared budget", err)
			}
			entries, err := os.ReadDir(destination)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Fatal("combined byte boundary checked after member creation")
			}
		}
	}
}

func TestArchiveExtractionOptionsRejectInvalidOverrides(t *testing.T) {
	for _, name := range []string{"BOSH_RELEASE_MAX_EXPANDED_BYTES", "BOSH_RELEASE_MAX_ARCHIVE_MEMBERS", "BOSH_RELEASE_MAX_MANIFEST_BYTES"} {
		for _, value := range []string{"0", "-1", "many"} {
			t.Setenv(name, value)
			if _, err := util.ArchiveExtractionOptions(t.TempDir()); err == nil || !strings.Contains(err.Error(), name) {
				t.Fatal(name, value, err)
			}
		}
		t.Setenv(name, "")
	}
}

func TestRelatedArchivesShareManifestBudget(t *testing.T) {
	t.Setenv("BOSH_RELEASE_MAX_EXPANDED_BYTES", "101")
	t.Setenv("BOSH_RELEASE_MAX_ARCHIVE_MEMBERS", "13")
	t.Setenv("BOSH_RELEASE_MAX_MANIFEST_BYTES", "3")
	for i := 0; i < 2; i++ {
		archive := fixtureArchive(t, []*tar.Header{{Name: "job.MF", Size: 2, Mode: 0600, Typeflag: tar.TypeReg}}, []string{"{}"}, true)
		destination := t.TempDir()
		options, err := util.ArchiveExtractionOptions(destination)
		if err != nil {
			t.Fatal(err)
		}
		err = boundedCompressor().DecompressFileToDir(archive, destination, options)
		if i == 0 && err != nil {
			t.Fatal(err)
		}
		if i == 1 && (err == nil || !strings.Contains(err.Error(), "combined release manifests")) {
			t.Fatal("manifest reads reset shared budget", err)
		}
	}
}
