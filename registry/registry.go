package registry

type registry map[string][]byte

type Registry interface {
	Save(string, []byte) bool
	Get(string) ([]byte, bool)
	Delete(string)
}

func NewRegistry() registry {
	return make(registry)
}

func (i registry) Save(key string, value []byte) bool {
	_, exists := i[key]
	i[key] = value

	return exists
}

func (i registry) Get(key string) ([]byte, bool) {
	value, exists := i[key]

	return value, exists
}

func (i registry) Delete(key string) {
	delete(i, key)
}
