package resource_test

import (
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	calculator "github.com/cloudfoundry/bosh-cli/v7/crypto"
	pkg "github.com/cloudfoundry/bosh-cli/v7/release/pkg"
	resource "github.com/cloudfoundry/bosh-cli/v7/release/resource"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
)

type replacingArchiveReader struct{ fs boshsys.FileSystem }

func (r replacingArchiveReader) OpenFile(filename string, flag int, mode os.FileMode) (boshsys.File, error) {
	file, err := r.fs.OpenFile(filename, flag, mode)
	return &replacingArchiveFile{File: file, path: filename}, err
}

type replacingArchiveFile struct {
	boshsys.File
	path     string
	replaced bool
}

func (f *replacingArchiveFile) Read(data []byte) (int, error) {
	n, err := f.File.Read(data)
	if n > 0 && !f.replaced {
		f.replaced = true
		// Replace the pathname after returning the authentic bytes from its original
		// descriptor. Any subsequent reopen would see different archive contents.
		if writeErr := os.WriteFile(f.path+".new", []byte("replacement archive"), 0600); writeErr != nil {
			return n, writeErr
		}
		if renameErr := os.Rename(f.path+".new", f.path); renameErr != nil {
			return n, renameErr
		}
	}
	return n, err
}
func TestRehashDoesNotReopenVerifiedArchive(t *testing.T) {
	for _, compiled := range []bool{false, true} {
		t.Run(fmt.Sprintf("compiled=%v", compiled), func(t *testing.T) {
			original := []byte("authentic archive")
			filename := filepath.Join(t.TempDir(), "archive")
			if err := os.WriteFile(filename, original, 0600); err != nil {
				t.Fatal(err)
			}
			fs := boshsys.NewOsFileSystem(boshlog.NewLogger(boshlog.LevelNone))
			digestCalculator := calculator.NewDigestCalculator(fs, []boshcrypto.Algorithm{boshcrypto.DigestAlgorithmSHA256})
			old := fmt.Sprintf("%x", sha1.Sum(original))
			want := fmt.Sprintf("sha256:%x", sha256.Sum256(original))
			var actual string
			if compiled {
				result, err := pkg.NewCompiledPackageWithArchive("name", "fp", "os", filename, old, nil).RehashWithCalculator(digestCalculator, replacingArchiveReader{fs})
				if err != nil {
					t.Fatal(err)
				}
				actual = result.ArchiveDigest()
			} else {
				result, err := resource.NewResourceWithBuiltArchive("name", "fp", filename, old).RehashWithCalculator(digestCalculator, replacingArchiveReader{fs})
				if err != nil {
					t.Fatal(err)
				}
				actual = result.ArchiveDigest()
			}
			if actual != want {
				t.Fatalf("accepted replacement digest: got %s want %s", actual, want)
			}
		})
	}
}
