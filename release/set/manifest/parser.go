package manifest

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	"gopkg.in/yaml.v2"

	biutil "github.com/cloudfoundry/bosh-init/common/util"
	boshtpl "github.com/cloudfoundry/bosh-init/director/template"
	birelmanifest "github.com/cloudfoundry/bosh-init/release/manifest"
)

type Parser interface {
	Parse(string, boshtpl.Variables) (Manifest, error)
}

type parser struct {
	fs        boshsys.FileSystem
	validator Validator

	logTag string
	logger boshlog.Logger
}

type manifest struct {
	Releases []birelmanifest.ReleaseRef
}

func NewParser(fs boshsys.FileSystem, logger boshlog.Logger, validator Validator) Parser {
	return &parser{
		fs:        fs,
		validator: validator,

		logTag: "releaseSetParser",
		logger: logger,
	}
}

func (p *parser) Parse(path string, vars boshtpl.Variables) (Manifest, error) {
	contents, err := p.fs.ReadFile(path)
	if err != nil {
		return Manifest{}, bosherr.WrapErrorf(err, "Reading file %s", path)
	}

	tpl := boshtpl.NewTemplate(contents)

	result, err := tpl.Evaluate(vars)
	if err != nil {
		return Manifest{}, bosherr.WrapErrorf(err, "Evaluating manifest")
	}

	bytes := result.Content()
	comboManifest := manifest{}

	err = yaml.Unmarshal(bytes, &comboManifest)
	if err != nil {
		return Manifest{}, bosherr.WrapError(err, "Unmarshalling release set manifest")
	}

	p.logger.Debug(p.logTag, "Parsed release set manifest: %#v", comboManifest)

	for i, releaseRef := range comboManifest.Releases {
		comboManifest.Releases[i].URL, err = biutil.AbsolutifyPath(path, releaseRef.URL, p.fs)
		if err != nil {
			return Manifest{}, bosherr.WrapErrorf(err, "Resolving release path '%s", releaseRef.URL)
		}
	}

	releaseSetManifest := Manifest{
		Releases: comboManifest.Releases,
	}

	err = p.validator.Validate(releaseSetManifest)
	if err != nil {
		return Manifest{}, bosherr.WrapError(err, "Validating release set manifest")
	}

	return releaseSetManifest, nil
}
