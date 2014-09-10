package release

type JobReader interface {
	Read() (Job, error)
}
