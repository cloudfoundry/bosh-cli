package director

import (
	"strings"
)

type SkipDrain struct {
	All   bool
	Slugs []PoolOrInstanceSlug
}

func (s SkipDrain) AsQueryValue() string {
	if s.All {
		return "*"
	}

	pieces := []string{}

	for _, slug := range s.Slugs {
		pieces = append(pieces, slug.String())
	}

	return strings.Join(pieces, ",")
}

func (s *SkipDrain) UnmarshalFlag(data string) error {
	if len(data) == 0 {
		*s = SkipDrain{All: true}
		return nil
	}

	sd := SkipDrain{}

	pieces := strings.Split(data, ",")

	for _, p := range pieces {
		slug, err := NewPoolOrInstanceSlugFromString(p)
		if err != nil {
			return err
		}

		sd.Slugs = append(sd.Slugs, slug)
	}

	*s = sd

	return nil
}
