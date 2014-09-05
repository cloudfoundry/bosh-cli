package deployment

type LocalDeployment struct {
	name       string
	properties map[string]interface{}
}

func NewLocalDeployment(name string, properties map[string]interface{}) LocalDeployment {
	return LocalDeployment{
		name:       name,
		properties: properties,
	}
}

func (l LocalDeployment) Name() string {
	return l.name
}

func (l LocalDeployment) Properties() map[string]interface{} {
	return l.properties
}
