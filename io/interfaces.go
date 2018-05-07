package io

import sysio "io"

type ReadSeekCloser interface {
	sysio.Seeker
	sysio.ReadCloser
}
