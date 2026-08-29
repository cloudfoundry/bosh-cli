package property

type Map map[string]Property

func (m *Map) UnmarshalYAML(unmarshal func(any) error) error {
	rawMap := map[any]any{}
	err := unmarshal(&rawMap)
	if err != nil {
		return err
	}

	*m, err = BuildMap(rawMap)
	if err != nil {
		return err
	}

	return nil
}
