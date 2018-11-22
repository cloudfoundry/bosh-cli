package takeoutfakes

import (
	"errors"
	"fmt"
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	"github.com/cloudfoundry/bosh-cli/takeout"
	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

type FakeUtensils struct {
	takeout.RealUtensils
	RetrieveMap map[boshdir.ManifestRelease]string
}

func (c FakeUtensils) Reset() {
	c.RetrieveMap = make(map[boshdir.ManifestRelease]string)
}

func (c FakeUtensils) RetrieveRelease(r boshdir.ManifestRelease, ui boshui.UI, localFileName string) (err error) {
	if val, ok := c.RetrieveMap[r]; ok {
		return errors.New(fmt.Sprintf("Release already in download map: %s", val))
	}
	c.RetrieveMap[r] = localFileName
	return
}
