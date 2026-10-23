package cmd_test

import (
	"testing"

	"github.com/cloudfoundry/bosh-cli/v7/cmd"
	"github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
	"github.com/cloudfoundry/bosh-cli/v7/release"
	"github.com/cloudfoundry/bosh-cli/v7/release/releasefakes"
	"github.com/cloudfoundry/bosh-cli/v7/releasedir"
)

func TestReleaseOutputComponentsCheckedBeforeArchiveWrite(t *testing.T) {
	for _, names := range [][2]string{{"../outside", "1"}, {"release", "../outside"}} {
		input := &releasefakes.FakeRelease{}
		input.NameReturns(names[0])
		input.VersionReturns(names[1])
		reader := &releasefakes.FakeReader{}
		reader.ReadReturns(input, nil)
		factory := func(opts.DirOrCWDArg) (release.Reader, releasedir.ReleaseDir) { return reader, nil }
		command := cmd.NewCreateReleaseCmd(factory, nil, nil, nil)
		options := opts.CreateReleaseOpts{}
		options.Args.Manifest.Path = "manifest"
		options.Tarball.ExpandedPath = "((name))-((version)).tgz"
		if _, err := command.Run(options); err == nil {
			t.Fatal("accepted invalid release output component")
		}
	}
}
