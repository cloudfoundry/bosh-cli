package takeout

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type Utensils interface {
	ReleaseDownloader
	OpFileGenerator
}

type ReleaseDownloader interface {
	RetrieveRelease(r boshdir.ManifestRelease, ui boshui.UI, localFileName string) (err error)
}

type OpFileGenerator interface {
	TakeOutRelease(r boshdir.ManifestRelease, ui boshui.UI, mirrorPrefix string) (entry OpEntry, err error)
}
