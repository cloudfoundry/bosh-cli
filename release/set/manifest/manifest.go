package manifest

import (
	birelmanifest "github.com/cloudfoundry/bosh-init/release/manifest"
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
