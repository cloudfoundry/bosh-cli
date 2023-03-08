package decompress

import (
	"io"

	"github.com/klauspost/compress/zlib"
)

type GZip struct{}

func (g GZip) Reader(src io.Reader) (io.ReadCloser, error) {
	return zlib.NewReader(src)
}

func (g GZip) Resetable() bool { return true }

func (g GZip) Reset(old, src io.Reader) error {
	return old.(zlib.Resetter).Reset(src, nil)
}
