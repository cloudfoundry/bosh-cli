package cyclonedxjson

import (
	"io"

	"github.com/CycloneDX/cyclonedx-go"

	"github.com/anchore/syft/syft/sbom"
	"github.com/cloudfoundry/bosh-cli/v7/sbom/formats/common/cyclonedxhelpers"
)

func encoder(output io.Writer, s sbom.SBOM) error {
	bom := cyclonedxhelpers.ToFormatModel(s)
	enc := cyclonedx.NewBOMEncoder(output, cyclonedx.BOMFileFormatJSON)
	enc.SetPretty(true)
	enc.SetEscapeHTML(false)
	err := enc.Encode(bom)
	return err
}
