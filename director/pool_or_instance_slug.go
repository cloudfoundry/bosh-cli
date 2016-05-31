package director

import (
	"fmt"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type PoolOrInstanceSlug struct {
	name      string
	indexOrID string // optional
}

func NewPoolOrInstanceSlug(name, indexOrID string) PoolOrInstanceSlug {
	if len(name) == 0 {
		panic("Expected pool or instance to specify non-empty name")
	}
	return PoolOrInstanceSlug{name: name, indexOrID: indexOrID}
}

func NewPoolOrInstanceSlugFromString(str string) (PoolOrInstanceSlug, error) {
	pieces := strings.Split(str, "/")
	if len(pieces) != 1 && len(pieces) != 2 {
		return PoolOrInstanceSlug{}, bosherr.Errorf(
			"Expected pool or instance '%s' to be in format 'name' or 'name/id-or-index'", str)
	}

	if len(pieces[0]) == 0 {
		return PoolOrInstanceSlug{}, bosherr.Errorf(
			"Expected pool or instance '%s' to specify non-empty name", str)
	}

	slug := PoolOrInstanceSlug{name: pieces[0]}

	if len(pieces) == 2 {
		if len(pieces[1]) == 0 {
			return PoolOrInstanceSlug{}, bosherr.Errorf(
				"Expected instance '%s' to specify non-empty ID or index", str)
		}

		slug.indexOrID = pieces[1]
	}

	return slug, nil
}

func (s PoolOrInstanceSlug) Name() string      { return s.name }
func (s PoolOrInstanceSlug) IndexOrID() string { return s.indexOrID }

func (s PoolOrInstanceSlug) String() string {
	if len(s.indexOrID) > 0 {
		return fmt.Sprintf("%s/%s", s.name, s.indexOrID)
	}
	return s.name
}
