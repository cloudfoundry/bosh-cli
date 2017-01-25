package main

import (
	"fmt"
	boshcrypto "github.com/cloudfoundry/bosh-utils/crypto"
	"github.com/jessevdk/go-flags"
	"os"
)

type opts struct {
	VerifyMultiDigestCommand MultiDigestCommand `command:"verify-multi-digest"`
	VersionFlag              func() error       `long:"version"`
}

func main() {
	o := opts{}
	o.VersionFlag = func() error {
		return &flags.Error{
			Type:    flags.ErrHelp,
			Message: fmt.Sprintf("version %s\n", VersionLabel),
		}
	}

	_, err := flags.Parse(&o)

	if typedErr, ok := err.(*flags.Error); ok {
		if typedErr.Type == flags.ErrHelp {
			err = nil
		}
	}

	if err != nil {
		os.Exit(1)
	}
}

type MultiDigestArgs struct {
	File   string
	Digest string
}

type MultiDigestCommand struct {
	Args MultiDigestArgs `positional-args:"yes"`
}

func (m MultiDigestCommand) Execute(args []string) error {
	multipleDigest := boshcrypto.MustParseMultipleDigest(m.Args.Digest)
	file, err := os.Open(m.Args.File)
	if err != nil {
		return err
	}
	return multipleDigest.Verify(file)
}
