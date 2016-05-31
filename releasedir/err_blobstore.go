package releasedir

// ErrBlobstore postpones returning an error until one of the actions are performed.
type ErrBlobstore struct {
	err error
}

func NewErrBlobstore(err error) ErrBlobstore {
	return ErrBlobstore{err: err}
}

func (b ErrBlobstore) Get(blobID, sha1 string) (string, error)    { return "", b.err }
func (b ErrBlobstore) Create(path string) (string, string, error) { return "", "", b.err }
func (b ErrBlobstore) CleanUp(path string) error                  { return b.err }
func (b ErrBlobstore) Delete(blobID string) error                 { return b.err }
func (b ErrBlobstore) Validate() error                            { return b.err }
