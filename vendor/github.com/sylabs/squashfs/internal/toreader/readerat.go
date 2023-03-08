package toreader

import "io"

type ReaderAt struct {
	d []byte
}

func NewReaderAt(r io.Reader) (ra ReaderAt, err error) {
	ra.d, err = io.ReadAll(r)
	return
}

func (r ReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if int(off) >= len(r.d) {
		return 0, io.EOF
	}
	n = copy(p, r.d[off:])
	if n != len(p) {
		err = io.EOF
	}
	return
}
