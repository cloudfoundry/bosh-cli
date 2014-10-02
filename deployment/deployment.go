package deployment

type ReleaseJobRef struct {
	Name    string
	Release string
}

type Job struct {
	Name      string
	Instances int
	Templates []ReleaseJobRef
}

type NetworkType string

const (
	Dynamic NetworkType = "dynamic"
)

type Network struct {
	Name string
	Type NetworkType
}

type Deployment struct {
	Name       string
	Properties map[string]interface{}
	Jobs       []Job
	Networks   []Network
}
