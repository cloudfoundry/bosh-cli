package director

import (
	"io"
)

type NoopFileReporter struct{}

func NewNoopFileReporter() NoopFileReporter {
	return NoopFileReporter{}
}

func (r NoopFileReporter) TrackUpload(size int64, reader io.ReadCloser) io.ReadCloser { return reader }
func (r NoopFileReporter) TrackDownload(size int64, writer io.Writer) io.Writer       { return writer }

type NoopTaskReporter struct{}

func NewNoopTaskReporter() NoopTaskReporter {
	return NoopTaskReporter{}
}

func (r NoopTaskReporter) TaskStarted(id int)                   {}
func (r NoopTaskReporter) TaskFinished(id int, state string)    {}
func (r NoopTaskReporter) TaskOutputChunk(id int, chunk []byte) {}
