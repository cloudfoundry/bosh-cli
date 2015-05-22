package manifest

import (
	birelmanifest "github.com/cloudfoundry/bosh-init/release/manifest"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Manifest struct {
	Releases []birelmanifest.ReleaseRef
}

func (d Manifest) ReleasesByName() map[string]birelmanifest.ReleaseRef {
	releasesByName := map[string]birelmanifest.ReleaseRef{}
	for _, release := range d.Releases {
		releasesByName[release.Name] = release
	}
	return releasesByName
}

func (d Manifest) FindByName(name string) (birelmanifest.ReleaseRef, bool) {
	for _, release := range d.Releases {
		if release.Name == name {
			return release, true
		}
	}
	return birelmanifest.ReleaseRef{}, false
}

func ParseAndValidateFrom(deploymentManifestPath string, parser Parser, validator Validator) (Manifest, error) {
	releaseSetManifest, err := parser.Parse(deploymentManifestPath)
	if err != nil {
		return Manifest{}, bosherr.WrapErrorf(err, "Parsing release set manifest '%s'", deploymentManifestPath)
	}

	err = validator.Validate(releaseSetManifest)
	if err != nil {
		return Manifest{}, bosherr.WrapError(err, "Validating release set manifest")
	}
	return releaseSetManifest, nil
}
