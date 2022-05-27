package ssh

import (
	"strings"

	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
)

type printableHost struct {
	boshdir.Host
}

func (h printableHost) String() string {
	if strings.Contains(h.Host.Host, ":") {
		return "[" + h.Host.Host + "]"
	}
	return h.Host.Host
}
