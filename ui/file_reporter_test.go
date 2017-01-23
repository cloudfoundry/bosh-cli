package ui_test

import (
	. "github.com/cloudfoundry/bosh-cli/ui"

	"github.com/cloudfoundry/bosh-cli/ui/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var seekCalls []interface{}

type FakeSeekableReader struct{}
type FakeReaderCloser struct{}

func (FakeSeekableReader) Read(p []byte) (n int, err error) {
	panic("should not call")
}

func (FakeSeekableReader) Close() error {
	panic("should not call")
}

func (FakeReaderCloser) Read(p []byte) (n int, err error) {
	panic("should not call")
}

func (FakeReaderCloser) Close() error {
	panic("should not call")
}

func (r FakeSeekableReader) Seek(offset int64, whence int) (int64, error) {
	seekCalls = append(seekCalls, []interface{}{offset, whence})

	return 0, nil
}

var _ = Describe("ReadCloserProxy", func() {
	BeforeEach(func() {
		seekCalls = make([]interface{}, 1)
	})
	Describe("Seek", func() {
		Context("when reader is seekable", func() {
			It("delegates to internal seeker", func() {
				seekerReader := FakeSeekableReader{}
				fileReporter := NewFileReporter(&fakes.FakeUI{})
				readCloserProxy := fileReporter.TrackUpload(0, seekerReader)

				readCloserProxy.Seek(12, 42)
				Expect(seekCalls).To(ContainElement([]interface{}{int64(12), 42}))
			})
		})

		Context("when reader is NOT seekable", func() {
			It("does not complain and returns 0, nil", func() {
				reader := FakeReaderCloser{}
				fileReporter := NewFileReporter(&fakes.FakeUI{})
				readCloserProxy := fileReporter.TrackUpload(0, reader)

				bytes, err := readCloserProxy.Seek(12, 42)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes).To(Equal(int64(0)))
			})
		})
	})
})
