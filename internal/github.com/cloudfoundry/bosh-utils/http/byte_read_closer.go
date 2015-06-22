package http

import (
	"bytes"
	"errors"
	"io"
	"strings"
)

type bytesReadCloser struct {
	reader io.Reader
	closed bool
}

func NewBytesReadCloser(content []byte) io.ReadCloser {
	return &bytesReadCloser{
		reader: bytes.NewReader(content),
		closed: false,
	}
}

func NewStringReadCloser(content string) io.ReadCloser {
	return &bytesReadCloser{
		reader: strings.NewReader(content),
		closed: false,
	}
}

func (s *bytesReadCloser) Close() error {
	s.closed = true
	return nil
}

func (s *bytesReadCloser) Closed() bool {
	return s.closed
}

func (s *bytesReadCloser) Read(p []byte) (n int, err error) {
	if s.closed {
		return 0, errors.New("already closed")
	}

	return s.reader.Read(p)
}
