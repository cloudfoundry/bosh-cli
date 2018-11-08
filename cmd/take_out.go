package cmd

import (
	"crypto/sha1"
	"fmt"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"gopkg.in/yaml.v2"
	"io"
	"net/http"
	"os"
	"regexp"
)

var BadChar = regexp.MustCompile("[?=\"]")

type OpEntry struct {
	Type  string `json:"type" yaml:"type"`
	Path  string `json:"path" yaml:"path"`
	Value string `json:"value" yaml:"value"`
}

type TakeOutCmd struct {
	ui boshui.UI
}

func NewTakeOutCmd(ui boshui.UI) TakeOutCmd {
	return TakeOutCmd{ui: ui}
}

func (c TakeOutCmd) Run(opts TakeOutOpts) error {
	tpl := boshtpl.NewTemplate(opts.Args.Manifest.Bytes)

	bytes, err := tpl.Evaluate(opts.VarFlags.AsVariables(), opts.OpsFlags.AsOp(), boshtpl.EvaluateOpts{})
	if err != nil {
		return bosherr.WrapErrorf(err, "Evaluating manifest")
	}
	if _, err := os.Stat(opts.Args.Name); os.IsExist(err) {
		return bosherr.WrapErrorf(err, "Takeout op name exists")
	}

	manifest, err := boshdir.NewManifestFromBytes(bytes)
	c.ui.PrintLinef("Processing releases for offline use")
	var releaseChanges []OpEntry
	for _, r := range manifest.Releases {
		o, err := TakeOutRelease(r, c.ui)
		if err != nil {
			return err
		}
		releaseChanges = append(releaseChanges, o)
	}

	y, _ := yaml.Marshal(releaseChanges)
	c.ui.PrintLinef("Writing take-out operation to file: " + opts.Args.Name)
	takeoutOp, err := os.Create(opts.Args.Name)
	if err != nil {
		return err
	}
	defer takeoutOp.Close()
	takeoutOp.WriteString("---\n")
	takeoutOp.WriteString(string(y))
	return nil
}
func TakeOutRelease(r boshdir.ManifestRelease, ui boshui.UI) (entry OpEntry, err error) {

	// generate a local file name that's safe
	localFileName := BadChar.ReplaceAllString(fmt.Sprintf("%s_v%s.tgz", r.Name, r.Version), "_")

	if _, err := os.Stat(localFileName); os.IsNotExist(err) {
		err = RetrieveRelease(r, ui, localFileName)
		if err != nil {
			return OpEntry{}, err
		}
	} else {
		ui.PrintLinef("Release present: %s / %s -> %s", r.Name, r.Version, localFileName)
	}
	if len(r.Name) > 0 {
		path := fmt.Sprintf("/releases/name=%s/url", r.Name)
		localFile := fmt.Sprintf("file://%s", localFileName)
		entry = OpEntry{Type: "replace", Path: path, Value: localFile}
	}
	return entry, err
}
func RetrieveRelease(r boshdir.ManifestRelease, ui boshui.UI, localFileName string) (err error) {
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
