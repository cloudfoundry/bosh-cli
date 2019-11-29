package takeout

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type Utensils interface {
	DeploymentReader
	ReleaseDownloader
	OpFileGenerator
	StemcellDownloader
}
type Manifest struct {
	Name string

	Releases  []boshdir.ManifestRelease
	Stemcells []boshdir.ManifestReleaseStemcell
}

type DeploymentReader interface {
	ParseDeployment(bytes []byte) (Manifest, error)
}

type ReleaseDownloader interface {
	RetrieveRelease(r boshdir.ManifestRelease, ui boshui.UI, localFileName string) (err error)
}

type StemcellDownloader interface {
	RetrieveStemcell(s boshdir.ManifestReleaseStemcell, ui boshui.UI, stemCellType string) (err error)
}

type OpFileGenerator interface {
	TakeOutRelease(r boshdir.ManifestRelease, ui boshui.UI, mirrorPrefix string) (entry OpEntry, err error)
}
