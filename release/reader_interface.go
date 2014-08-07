package release

type Reader interface {
	Read() (Release, error)
}
