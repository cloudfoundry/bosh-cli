package takeout

import (
	"crypto/sha1"
	"fmt"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"io"
	"net/http"
	"os"
	"regexp"
)

var BadChar = regexp.MustCompile("[?=\"]")

type RealUtensils struct {
}

func (c RealUtensils) RetrieveRelease(r boshdir.ManifestRelease, ui boshui.UI, localFileName string) (err error) {
	ui.PrintLinef("Downloading release: %s / %s -> %s", r.Name, r.Version, localFileName)

	resp, err := http.Get(r.URL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(localFileName)
	if err != nil {
		return
	}
	defer out.Close()

	// Write the body to file
	hash := sha1.New()
	_, err = io.Copy(out, io.TeeReader(resp.Body, hash))
	actualSha1 := fmt.Sprintf("%x", hash.Sum(nil))
	if err != nil {
		return
	}
	if len(r.SHA1) == 40 {
		if actualSha1 != r.SHA1 {
			return bosherr.Errorf("sha1 mismatch %s (a:%s, e:%s)", localFileName, actualSha1, r.SHA1)
		}
	}
	return
}

func (c RealUtensils) TakeOutRelease(r boshdir.ManifestRelease, ui boshui.UI, mirrorPrefix string) (entry OpEntry, err error) {

	// generate a local file name that's safe
	localFileName := BadChar.ReplaceAllString(fmt.Sprintf("%s_v%s.tgz", r.Name, r.Version), "_")

	if _, err := os.Stat(localFileName); os.IsNotExist(err) {
		err = c.RetrieveRelease(r, ui, localFileName)
		if err != nil {
			return OpEntry{}, err
		}
	} else {
		ui.PrintLinef("Release present: %s / %s -> %s", r.Name, r.Version, localFileName)
	}
	if len(r.Name) > 0 {
		path := fmt.Sprintf("/releases/name=%s/url", r.Name)
		localFile := fmt.Sprintf("%s%s", mirrorPrefix, localFileName)
		entry = OpEntry{Type: "replace", Path: path, Value: localFile}
	}
	return entry, err
}

func TakeOutUtensils() (utensils Utensils) {
	return RealUtensils{}
}
