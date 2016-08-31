package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshfu "github.com/cloudfoundry/bosh-utils/fileutil"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"github.com/pivotal-golang/clock"

	bicrypto "github.com/cloudfoundry/bosh-cli/crypto"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	biui "github.com/cloudfoundry/bosh-cli/ui"
)

type Downloader interface {
	Download(blobstoreID, sha1, prefix, dstDirPath string) error
}

type UIDownloader struct {
	director    boshdir.Director
	sha1calc    bicrypto.SHA1Calculator
	timeService clock.Clock

	fs boshsys.FileSystem
	ui biui.UI
}

func NewUIDownloader(
	director boshdir.Director,
	sha1calc bicrypto.SHA1Calculator,
	timeService clock.Clock,
	fs boshsys.FileSystem,
	ui biui.UI,
) UIDownloader {
	return UIDownloader{
		director:    director,
		sha1calc:    sha1calc,
		timeService: timeService,

		fs: fs,
		ui: ui,
	}
}

func (d UIDownloader) Download(blobstoreID, sha1, prefix, dstDirPath string) error {
	tsSuffix := strings.Replace(d.timeService.Now().Format("20060102-150405.999999999"), ".", "-", -1)

	dstFileName := fmt.Sprintf("%s-%s.tgz", prefix, tsSuffix)

	dstFilePath := filepath.Join(dstDirPath, dstFileName)

	tmpFile, err := d.fs.TempFile(fmt.Sprintf("director-resource-%s", blobstoreID))
	if err != nil {
		return err
	}

	defer d.fs.RemoveAll(tmpFile.Name())

	d.ui.PrintLinef("Downloading resource '%s' to '%s'...", blobstoreID, dstFilePath)

	err = d.director.DownloadResourceUnchecked(blobstoreID, tmpFile)
	if err != nil {
		return err
	}

	actualSHA1, err := d.sha1calc.Calculate(tmpFile.Name())
	if err != nil {
		return err
	}

	if len(sha1) > 0 && sha1 != actualSHA1 {
		return bosherr.Errorf("Expected file SHA1 to be '%s' but was '%s'", sha1, actualSHA1)
	}

	err = boshfu.NewFileMover(d.fs).Move(tmpFile.Name(), dstFilePath)
	if err != nil {
		return bosherr.WrapErrorf(err, "Moving to final destination")
	}

	return nil
}
