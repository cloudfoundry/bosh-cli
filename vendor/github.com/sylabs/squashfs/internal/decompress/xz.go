package decompress

import (
	"io"

	"github.com/therootcompany/xz"
)

type Xz struct{}

func (x Xz) Reader(r io.Reader) (io.ReadCloser, error) {
	rdr, err := xz.NewReader(r, 0)
	return io.NopCloser(rdr), err
}

func (x Xz) Resetable() bool { return true }

func (x Xz) Reset(old, src io.Reader) error {
	return old.(*xz.Reader).Reset(src)
}
