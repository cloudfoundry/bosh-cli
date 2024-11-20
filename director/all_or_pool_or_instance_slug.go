package director

import (
	"fmt"
	"net"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type AllOrInstanceGroupOrInstanceSlug struct {
	name      string // optional
	indexOrID string // optional
	ip        string // optional
}

func NewAllOrInstanceGroupOrInstanceSlug(name, indexOrID string) AllOrInstanceGroupOrInstanceSlug {
	return AllOrInstanceGroupOrInstanceSlug{name: name, indexOrID: indexOrID}
}

func NewAllOrInstanceGroupOrInstanceSlugFromString(str string) (AllOrInstanceGroupOrInstanceSlug, error) {
	return parseAllOrInstanceGroupOrInstanceSlug(str)
}

func (s AllOrInstanceGroupOrInstanceSlug) Name() string      { return s.name }
func (s AllOrInstanceGroupOrInstanceSlug) IndexOrID() string { return s.indexOrID }
func (s AllOrInstanceGroupOrInstanceSlug) IP() string        { return s.ip }

func (s AllOrInstanceGroupOrInstanceSlug) InstanceSlug() (InstanceSlug, bool) {
	if len(s.name) > 0 && len(s.indexOrID) > 0 {
		return NewInstanceSlug(s.name, s.indexOrID), true
	}
	return InstanceSlug{}, false
}

func (s AllOrInstanceGroupOrInstanceSlug) Overlaps(other AllOrInstanceGroupOrInstanceSlug) bool {
	if s.name != other.name {
		return false
	}

	// There is an overlap if either slug is empty or if the indexOrID or IP matches
	return s.indexOrID == "" || other.indexOrID == "" || s.indexOrID == other.indexOrID || s.ip == other.ip

}

func (s AllOrInstanceGroupOrInstanceSlug) String() string {
	if len(s.indexOrID) > 0 {
		return fmt.Sprintf("%s/%s", s.name, s.indexOrID)
	}
	return s.name
}

func (s *AllOrInstanceGroupOrInstanceSlug) UnmarshalFlag(data string) error {
	slug, err := parseAllOrInstanceGroupOrInstanceSlug(data)
	if err != nil {
		return err
	}

	*s = slug

	return nil
}

func parseAllOrInstanceGroupOrInstanceSlug(str string) (AllOrInstanceGroupOrInstanceSlug, error) {
	if len(str) == 0 {
		return AllOrInstanceGroupOrInstanceSlug{}, nil
	}

	ip := net.ParseIP(str)
	if ip != nil {
		return AllOrInstanceGroupOrInstanceSlug{ip: str}, nil
	}

	pieces := strings.Split(str, "/")
	if len(pieces) != 1 && len(pieces) != 2 {
		return AllOrInstanceGroupOrInstanceSlug{}, bosherr.Errorf(
			"Expected pool or instance '%s' to be in format 'name' or 'name/id-or-index'", str)
	}

	if len(pieces[0]) == 0 {
		return AllOrInstanceGroupOrInstanceSlug{}, bosherr.Errorf(
			"Expected pool or instance '%s' to specify non-empty name", str)
	}

	slug := AllOrInstanceGroupOrInstanceSlug{name: pieces[0]}

	if len(pieces) == 2 {
		if len(pieces[1]) == 0 {
			return AllOrInstanceGroupOrInstanceSlug{}, bosherr.Errorf(
				"Expected instance '%s' to specify non-empty ID or index", str)
		}

		slug.indexOrID = pieces[1]
	}

	return slug, nil
}
