package blobextract_test

import (
	"os"
	"testing"

	"github.com/cloudfoundry/bosh-cli/v7/installation/blobextract"
	blobfakes "github.com/cloudfoundry/bosh-utils/blobstore/blobstorefakes"
	compressorfakes "github.com/cloudfoundry/bosh-utils/fileutil/fakes"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	fsfakes "github.com/cloudfoundry/bosh-utils/system/fakes"
)

func TestExtractDirectoryModeBeforeUnpacking(t *testing.T) {
	for _, existing := range []bool{false, true} {
		fs := fsfakes.NewFakeFileSystem()
		if existing {
			if err := fs.MkdirAll("target", 0777); err != nil {
				t.Fatal(err)
			}
		}
		blobs := &blobfakes.FakeDigestBlobstore{}
		blobs.GetReturns("archive", nil)
		compressor := compressorfakes.NewFakeCompressor()
		compressor.DecompressFileToDirCallBack = func() {
			info, err := fs.Stat("target")
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode()&os.ModePerm != 0750 {
				t.Fatalf("directory mode before extraction %o", info.Mode()&os.ModePerm)
			}
		}
		extractor := blobextract.NewExtractor(fs, compressor, blobs, boshlog.NewLogger(boshlog.LevelNone))
		if err := extractor.Extract("blob", "fakesha1", "target"); err != nil {
			t.Fatal(err)
		}
	}
}
