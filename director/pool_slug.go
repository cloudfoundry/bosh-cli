package director

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type PoolSlug struct {
	name string
}

func NewPoolSlug(name string) PoolSlug {
	if len(name) == 0 {
		panic("Expected non-empty pool name")
	}
	return PoolSlug{name: name}
}

func (s PoolSlug) Name() string   { return s.name }
func (s PoolSlug) String() string { return s.name }

func (s *PoolSlug) UnmarshalFlag(data string) error {
	slug, err := parsePoolSlug(data)
	if err != nil {
		return err
	}

	*s = slug

	return nil
}

func parsePoolSlug(str string) (PoolSlug, error) {
	if len(str) == 0 {
		return PoolSlug{}, bosherr.Error("Expected non-empty pool name")
	}

	return PoolSlug{name: str}, nil
}
