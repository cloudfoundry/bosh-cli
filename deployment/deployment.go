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

type Deployment struct {
	Name       string
	Properties map[string]interface{}
	Jobs       []Job
}
