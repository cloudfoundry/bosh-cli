package tar

type Extractor interface {
	Extract(source string, destination string) error
}
