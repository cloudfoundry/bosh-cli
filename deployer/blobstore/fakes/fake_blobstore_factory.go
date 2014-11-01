package fakes

import (
	bmblobstore "github.com/cloudfoundry/bosh-micro-cli/deployer/blobstore"
)

type FakeBlobstoreFactory struct {
	CreateBlobstoreURL string
	CreateBlobstore    bmblobstore.Blobstore
	CreateErr          error
}

func NewFakeBlobstoreFactory() *FakeBlobstoreFactory {
	return &FakeBlobstoreFactory{}
}

func (f *FakeBlobstoreFactory) Create(blobstoreURL string) (bmblobstore.Blobstore, error) {
	f.CreateBlobstoreURL = blobstoreURL
	return f.CreateBlobstore, f.CreateErr
}
