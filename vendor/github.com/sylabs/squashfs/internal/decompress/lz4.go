package decompress

import (
	"io"

	"github.com/pierrec/lz4/v4"
)

type Lz4 struct{}

func (l Lz4) Reader(r io.Reader) (io.ReadCloser, error) {
	return io.NopCloser(lz4.NewReader(r)), nil
}

func (l Lz4) Resetable() bool { return true }

func (l Lz4) Reset(old, src io.Reader) error {
	old.(*lz4.Reader).Reset(src)
	return nil
}
