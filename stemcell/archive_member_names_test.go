package stemcell_test

import (
	"path/filepath"
	"testing"

	compressorfakes "github.com/cloudfoundry/bosh-utils/fileutil/fakes"
	fsfakes "github.com/cloudfoundry/bosh-utils/system/fakes"

	"github.com/cloudfoundry/bosh-cli/v7/stemcell"
)

func TestStemcellMemberNamesCannotBeOptions(t *testing.T) {
	for _, name := range []string{"--checkpoint=1", "--exclude=*", "-Icommand"} {
		fs := fsfakes.NewFakeFileSystem()
		compressor := compressorfakes.NewFakeCompressor()
		fs.SetGlob(filepath.Join("extracted", "*"), []string{filepath.Join("extracted", name)})
		input := stemcell.NewExtractedStemcell(stemcell.Manifest{}, "extracted", compressor, fs)
		if err := input.Pack("output.tgz"); err == nil {
			t.Fatalf("accepted archive option %q", name)
		}
		if compressor.CompressSpecificFilesInDirDir != "" {
			t.Fatal("passed option-like member to tar")
		}
	}
}
