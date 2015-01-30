package acceptance

import (
	"io"
)

// MultiWriter implements mirrored writes to multiple io.Writer objects.
type MultiWriter struct {
	err     error
	n       int
	writers []io.Writer
}

// NewMultiWriter returns a new MultiWriter that proxies writes to multiple io.Writer objects.
func NewMultiWriter(writers ...io.Writer) *MultiWriter {
	return &MultiWriter{
		writers: writers,
	}
}

// Write writes the contents of p into all proxied writers.
// It returns the number of bytes written.
// If nn < len(p), it also returns an error explaining
// why the write is short.
// If a proxied write errors, subsequent writers will not be written to.
func (b *MultiWriter) Write(p []byte) (nn int, err error) {
	for _, writer := range b.writers {
		nn, err = writer.Write(p)
		if err != nil {
			return nn, err
		}
	}
	return nn, nil
}
