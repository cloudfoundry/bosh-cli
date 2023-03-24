package decompress

import (
	"errors"
	"io"
)

var ErrNotResetable = errors.New("decompressor not resetable")

type Decompressor interface {
	//Creates a new decompressor reading from src.
	Reader(src io.Reader) (io.ReadCloser, error)
	//Reports whether Reset will work or not.
	Resetable() bool
	//Reset attempts to re-use an old decompressor with new data.
	//Will return ErrNotResetable if not Resetable().
	//Must ALWAYS be provided with a reader created with Reader.
	Reset(old, src io.Reader) error
}
