package crypto

import "io"

type Digest interface {
	Verify(io.Reader) error
	Algorithm() Algorithm
	String() string
}

var _ Digest = digestImpl{}

type Algorithm interface {
	CreateDigest(io.Reader) (Digest, error)
	Name() string
}

var _ Algorithm = algorithmSHAImpl{}
var _ Algorithm = unknownAlgorithmImpl{}
