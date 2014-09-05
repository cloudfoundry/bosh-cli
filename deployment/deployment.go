package deployment

type Deployment interface {
	Name() string
	Properties() map[string]interface{}
}
