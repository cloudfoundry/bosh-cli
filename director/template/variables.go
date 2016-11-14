package template

type Variables interface {
	Get(VariableDefinition) (interface{}, bool, error)
}

type VariableDefinition struct {
	Name    string
	Type    string
	Options interface{}
}
