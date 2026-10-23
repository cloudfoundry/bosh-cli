package util_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/bosh-cli/v7/common/util"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/cloudfoundry/bosh-utils/system"
)

type checkedFS struct {
	system.FileSystem
	t *testing.T
}
type checkedFile struct {
	system.File
	t *testing.T
}

func (fs checkedFS) OpenFile(name string, flags int, mode os.FileMode) (system.File, error) {
	if flags&os.O_CREATE != 0 && mode.Perm()&0077 != 0 {
		fs.t.Fatalf("new file exposed with mode %o", mode)
	}
	f, err := fs.FileSystem.OpenFile(name, flags, mode)
	if err != nil {
		return nil, err
	}
	return checkedFile{f, fs.t}, nil
}
func (f checkedFile) Write(b []byte) (int, error) {
	info, err := f.Stat()
	if err != nil {
		f.t.Fatal(err)
	}
	if info.Mode().Perm()&0077 != 0 {
		f.t.Fatalf("file has mode %o when writing contents", info.Mode().Perm())
	}
	return f.File.Write(b)
}
func TestPrivateFileModeBeforeWrite(t *testing.T) {
	fs := checkedFS{system.NewOsFileSystem(boshlog.NewLogger(boshlog.LevelNone)), t}
	name := filepath.Join(t.TempDir(), "nested", "config")
	for _, exists := range []bool{false, true} {
		if exists {
			if err := os.Chmod(name, 0644); err != nil {
				t.Fatal(err)
			}
		}
		if err := util.WritePrivateFile(fs, name, []byte("private-value")); err != nil {
			t.Fatal(err)
		}
		contents, err := os.ReadFile(name)
		if err != nil || string(contents) != "private-value" {
			t.Fatalf("stored content: %s %v", contents, err)
		}
	}
}
