package cyclonedxhelpers

import (
	"github.com/CycloneDX/cyclonedx-go"

	"github.com/anchore/syft/syft/pkg"
)

func encodeLicenses(p pkg.Package) *cyclonedx.Licenses {
	return nil
}

func decodeLicenses(c *cyclonedx.Component) (out []string) {
	if c.Licenses != nil {
		for _, l := range *c.Licenses {
			if l.License != nil {
				var lic string
				switch {
				case l.License.ID != "":
					lic = l.License.ID
				case l.License.Name != "":
					lic = l.License.Name
				default:
					continue
				}
				out = append(out, lic)
			}
		}
	}
	return
}
