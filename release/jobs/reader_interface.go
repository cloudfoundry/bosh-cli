package jobs

type Reader interface {
	Read() (Job, error)
}
