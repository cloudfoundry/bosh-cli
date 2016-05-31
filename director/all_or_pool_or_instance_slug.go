package director

import (
	"fmt"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type AllOrPoolOrInstanceSlug struct {
	name      string // optional
	indexOrID string // optional
}

func NewAllOrPoolOrInstanceSlug(name, indexOrID string) AllOrPoolOrInstanceSlug {
	return AllOrPoolOrInstanceSlug{name: name, indexOrID: indexOrID}
}

func NewAllOrPoolOrInstanceSlugFromString(str string) (AllOrPoolOrInstanceSlug, error) {
	return parseAllOrPoolOrInstanceSlug(str)
}

func (s AllOrPoolOrInstanceSlug) Name() string      { return s.name }
func (s AllOrPoolOrInstanceSlug) IndexOrID() string { return s.indexOrID }

func (s AllOrPoolOrInstanceSlug) InstanceSlug() (InstanceSlug, bool) {
	if len(s.name) > 0 && len(s.indexOrID) > 0 {
		return NewInstanceSlug(s.name, s.indexOrID), true
	}
	return InstanceSlug{}, false
}

func (s AllOrPoolOrInstanceSlug) String() string {
	if len(s.indexOrID) > 0 {
		return fmt.Sprintf("%s/%s", s.name, s.indexOrID)
	}
	return s.name
}

func (s *AllOrPoolOrInstanceSlug) UnmarshalFlag(data string) error {
	slug, err := parseAllOrPoolOrInstanceSlug(data)
	if err != nil {
		return err
	}

	*s = slug

	return nil
}

func parseAllOrPoolOrInstanceSlug(str string) (AllOrPoolOrInstanceSlug, error) {
	if len(str) == 0 {
		return AllOrPoolOrInstanceSlug{}, nil
	}

	pieces := strings.Split(str, "/")
	if len(pieces) != 1 && len(pieces) != 2 {
		return AllOrPoolOrInstanceSlug{}, bosherr.Errorf(
			"Expected pool or instance '%s' to be in format 'name' or 'name/id-or-index'", str)
	}

	if len(pieces[0]) == 0 {
		return AllOrPoolOrInstanceSlug{}, bosherr.Errorf(
			"Expected pool or instance '%s' to specify non-empty name", str)
	}

	slug := AllOrPoolOrInstanceSlug{name: pieces[0]}

	if len(pieces) == 2 {
		if len(pieces[1]) == 0 {
			return AllOrPoolOrInstanceSlug{}, bosherr.Errorf(
				"Expected instance '%s' to specify non-empty ID or index", str)
		}

		slug.indexOrID = pieces[1]
	}

	return slug, nil
}
