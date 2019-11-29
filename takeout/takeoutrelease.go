package takeout

import (
	"crypto/sha1"
	"fmt"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"gopkg.in/yaml.v2"
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

	tempFileName := localFileName + ".download"
	resp, err := http.Get(r.URL)

	if resp != nil {
		defer func() {
			if ferr := resp.Body.Close(); ferr != nil {
				err = ferr
			}
		}()
	}
	if err != nil {
		return err
	}

	// Create the file
	out, err := os.Create(tempFileName)
	if out != nil {
		defer func() {
			if ferr := out.Close(); ferr != nil {
				err = ferr
			}
		}()
	}
	if err != nil {
		return err
	}

	// Write the body to file
	hash := sha1.New()
	_, err = io.Copy(out, io.TeeReader(resp.Body, hash))
	actualSha1 := fmt.Sprintf("%x", hash.Sum(nil))
	if err != nil {
		return err
	}
	if len(r.SHA1) == 40 {
		if actualSha1 != r.SHA1 {
			return bosherr.Errorf("sha1 mismatch %s (a:%s, e:%s)", localFileName, actualSha1, r.SHA1)
		}
	}
	err = os.Rename(tempFileName, localFileName)
	if err != nil {
		return err
	}
	return nil
}

func (c RealUtensils) RetrieveStemcell(s boshdir.ManifestReleaseStemcell, ui boshui.UI, stemCellType string) (err error) {

	localFileName := fmt.Sprintf("bosh-%s-%s-go_agent-stemcell_v%s.tgz", stemCellType, s.OS, s.Version)
	ui.PrintLinef("Downloading stemcell: %s / %s -> %s", s.OS, s.Version, localFileName)

	url := fmt.Sprintf("https://bosh.io/d/stemcells/bosh-%s-%s-go_agent?v=%s", stemCellType, s.OS, s.Version)

	ui.PrintLinef("Trying %s", url)

	resp, err := http.Get(url)
	if resp != nil {
		defer func() {
			if ferr := resp.Body.Close(); ferr != nil {
				err = ferr
			}
		}()
	}
	if err != nil {
		return
	}

	// Create the file
	out, err := os.Create(localFileName)
	if err != nil {
		return
	}
	defer func() {
		if ferr := out.Close(); ferr != nil {
			err = ferr
		}
	}()

	// Write the body to file
	hash := sha1.New()
	_, err = io.Copy(out, io.TeeReader(resp.Body, hash))
	actualSha1 := fmt.Sprintf("%x", hash.Sum(nil))
	if err != nil {
		return
	}
	ui.PrintLinef("Stemcell %s SHA1:%s", localFileName, actualSha1)
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

func (c RealUtensils) ParseDeployment(bytes []byte) (Manifest, error) {
	var deployment Manifest

	err := yaml.Unmarshal(bytes, &deployment)
	if err != nil {
		return deployment, err
	}

	return deployment, nil
}
