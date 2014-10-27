package fakes

type FakeBlobstore struct {
	GetInputs []GetInput
	GetErr    error

	SaveInputs []SaveInput
	SaveErr    error
}

type GetInput struct {
	BlobID          string
	DestinationPath string
}

type SaveInput struct {
	BlobID     string
	SourcePath string
}

func NewFakeBlobstore() *FakeBlobstore {
	return &FakeBlobstore{}
}

func (b *FakeBlobstore) Get(blobID string, destinationPath string) error {
	b.GetInputs = append(b.GetInputs, GetInput{
		BlobID:          blobID,
		DestinationPath: destinationPath,
	})

	return b.GetErr
}

func (b *FakeBlobstore) Save(sourcePath string, blobID string) error {
	b.SaveInputs = append(b.SaveInputs, SaveInput{
		BlobID:     blobID,
		SourcePath: sourcePath,
	})

	return b.SaveErr
}
