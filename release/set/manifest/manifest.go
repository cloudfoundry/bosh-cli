package manifest

import (
	bmrelmanifest "github.com/cloudfoundry/bosh-micro-cli/release/manifest"
)

type Manifest struct {
	Releases []bmrelmanifest.ReleaseRef
}

func (d Manifest) ReleasesByName() map[string]bmrelmanifest.ReleaseRef {
	releasesByName := map[string]bmrelmanifest.ReleaseRef{}
	for _, release := range d.Releases {
		releasesByName[release.Name] = release
	}
	return releasesByName
}
