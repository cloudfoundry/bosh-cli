package manifest

import (
	biutil "github.com/cloudfoundry/bosh-init/common/util"
	bosherr "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-init/internal/github.com/cloudfoundry/bosh-utils/system"
	"github.com/cloudfoundry/bosh-init/internal/gopkg.in/yaml.v2"
	birelmanifest "github.com/cloudfoundry/bosh-init/release/manifest"
)

type Parser interface {
	Parse(path string) (Manifest, error)
}

type parser struct {
	fs        boshsys.FileSystem
	logger    boshlog.Logger
	logTag    string
	validator Validator
}

type manifest struct {
	Releases []birelmanifest.ReleaseRef
}

func NewParser(fs boshsys.FileSystem, logger boshlog.Logger, validator Validator) Parser {
	return &parser{
		fs:        fs,
		logger:    logger,
		logTag:    "releaseSetParser",
		validator: validator,
	}
}

func (p *parser) Parse(path string) (Manifest, error) {
	contents, err := p.fs.ReadFile(path)
	if err != nil {
		return Manifest{}, bosherr.WrapErrorf(err, "Reading file %s", path)
	}

	comboManifest := manifest{}
	err = yaml.Unmarshal(contents, &comboManifest)
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
