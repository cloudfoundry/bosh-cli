package configserver

type Client interface {
	Read(name string) (interface{}, error)
	Exists(name string) (bool, error)
	Write(name string, value interface{}) error
	Delete(name string) error
	Generate(name, type_ string, params interface{}) (interface{}, error)
}
