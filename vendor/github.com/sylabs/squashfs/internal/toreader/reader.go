package toreader

import "io"

type Reader struct {
	r   io.ReaderAt
	off int64
}

func NewReader(r io.ReaderAt, start int64) *Reader {
	return &Reader{
		r:   r,
		off: start,
	}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, err = r.r.ReadAt(p, r.off)
	r.off += int64(n)
	return
}

func (r Reader) Offset() int64 {
	return r.off
}
