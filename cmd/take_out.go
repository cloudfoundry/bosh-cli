package cmd

import (
	"fmt"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshtpl "github.com/cloudfoundry/bosh-cli/director/template"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"gopkg.in/yaml.v2"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
)
var BadChar = regexp.MustCompile("[?=]")

type OpEntry struct {
	Type string `json:"type" yaml:"type"`
	Path string `json:"path" yaml:"path"`
	Value string `json:"value" yaml:"value"`
}

type TakeOutCmd struct {
	ui              boshui.UI
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
	fmt.Println("Processing releases for offline use")
	var releaseChanges []OpEntry
	for _, r := range manifest.Releases {
		o, _ := DownloadFile(r)
		releaseChanges = append(releaseChanges, o)
	}

	y, _ := yaml.Marshal(releaseChanges)
	fmt.Println("Writing take-out operation")
	takeoutOp, err := os.Create(opts.Args.Name)
	if err != nil{
		return err
	}
	defer takeoutOp.Close()
	takeoutOp.WriteString(string(y))
	return nil
}
func DownloadFile(r boshdir.ManifestRelease) (entry OpEntry, err error) {

	// generate a local file name
	filepath := BadChar.ReplaceAllString(path.Base(r.URL),"_")

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		fmt.Println("Downloading release: " + r.Name + " " + r.Version)

		resp, err := http.Get(r.URL)
		if err != nil {
			return entry, err
		}
		defer resp.Body.Close()

		// Create the file
		out, err := os.Create(filepath)
		if err != nil {
			return entry, err
		}
		defer out.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return entry, err
		}
	} else {
		fmt.Println("Release already downloaded: " + r.Name + " " + r.Version)
	}
	path := fmt.Sprintf("/releases/name=%s/url", r.Name)
	localFile := fmt.Sprintf("file://%s", filepath)
	entry = OpEntry{Type: "replace", Path: path, Value: localFile}

	return entry, err
}

