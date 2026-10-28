package fileutil

import (
	"fmt"
	"io"
	"math"
	"sync"
)

// ArchiveBudget bounds the combined resources used by related nested archives.
type ArchiveBudget struct {
	mu                sync.Mutex
	payloadRemaining  int64
	streamRemaining   int64
	membersRemaining  int
	metadataRemaining int64
}

func NewArchiveBudget(bytes int64, members int, metadata ...int64) (*ArchiveBudget, error) {
	if bytes <= 0 || members <= 0 || int64(members) > (math.MaxInt64-1024*1024)/2048 {
		return nil, fmt.Errorf("invalid archive limits")
	}
	overhead := int64(members)*2048 + 1024*1024
	if bytes > math.MaxInt64-overhead-1 {
		return nil, fmt.Errorf("archive limits are too large")
	}
	metadataLimit := int64(32 * 1024 * 1024)
	if len(metadata) > 0 && metadata[0] > 0 {
		metadataLimit = metadata[0]
	}
	return &ArchiveBudget{payloadRemaining: bytes, streamRemaining: bytes + overhead + 1, membersRemaining: members, metadataRemaining: metadataLimit}, nil
}
func (b *ArchiveBudget) reserve(size int64) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.membersRemaining == 0 {
		return fmt.Errorf("combined archives exceed member budget")
	}
	if size < 0 || size > b.payloadRemaining {
		return fmt.Errorf("combined archives exceed expanded bytes budget")
	}
	b.membersRemaining--
	b.payloadRemaining -= size
	return nil
}

type budgetReader struct {
	reader io.Reader
	budget *ArchiveBudget
}

func (r budgetReader) Read(p []byte) (int, error) {
	b := r.budget
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.streamRemaining == 0 {
		return 0, fmt.Errorf("combined archive expanded stream exceeds byte/metadata budget")
	}
	if int64(len(p)) > b.streamRemaining {
		p = p[:int(b.streamRemaining)]
	}
	n, err := r.reader.Read(p)
	b.streamRemaining -= int64(n)
	return n, err
}

func (b *ArchiveBudget) reserveMetadata(size int64) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if size < 0 || size > b.metadataRemaining {
		return fmt.Errorf("combined release manifests exceed metadata byte budget")
	}
	b.metadataRemaining -= size
	return nil
}
