package cmd

import (
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts" //nolint:staticcheck
	boshreldir "github.com/cloudfoundry/bosh-cli/v7/releasedir"
)

type GenerateJobCmd struct {
	releaseDir boshreldir.ReleaseDir
}

func NewGenerateJobCmd(releaseDir boshreldir.ReleaseDir) GenerateJobCmd {
	return GenerateJobCmd{releaseDir: releaseDir}
}

func (c GenerateJobCmd) Run(opts GenerateJobOpts) error {
	templates := map[string]string{}

	for _, template := range opts.Templates {
		src, dest, ok := splitTemplateMapping(template)
		if !ok {
			return bosherr.Errorf("Expected template as SRC:DEST, got '%s'", template)
		}

		templates[src] = dest
	}

	return c.releaseDir.GenerateJob(boshreldir.GenerateJobOptions{
		Name:      opts.Args.Name,
		Monit:     opts.Monit,
		Templates: templates,
		Packages:  opts.Packages,
	})
}

// splitTemplateMapping splits SRC:DEST without mistaking a Windows drive
// letter for the separator. It splits on the first colon after the drive
// prefix so destinations may still contain colons.
func splitTemplateMapping(template string) (string, string, bool) {
	start := 0

	if len(template) > 2 && isDriveLetter(template[0]) && template[1] == ':' &&
		(template[2] == '\\' || template[2] == '/') {
		start = 2
	}

	separator := strings.Index(template[start:], ":")

	if separator < 0 {
		return "", "", false
	}

	separator += start

	if separator == 0 || separator == len(template)-1 {
		return "", "", false
	}

	return template[:separator], template[separator+1:], true
}

func isDriveLetter(ch byte) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z'
}
