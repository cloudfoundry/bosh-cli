## Development Workstation Setup

Note: This guide assumes a few things:

- You have gcc (or an equivalent)
- You can install packages (brew, apt-get, or equivalent)

Get Golang and its dependencies (Mac example, replace with your package manager of choice):

- `brew update`
- `brew install go`
- `brew install git` (Go needs git for the `go get` command)
- `brew install hg` (Go needs mercurial for the `go get` command)

Clone and set up the BOSH Micro CLI repository:

- `go get -d github.com/cloudfoundry/bosh-init`
- `cd $GOPATH/src/github.com/cloudfoundry/bosh-init`

From here on out we assume you're working in `$GOPATH/src/github.com/cloudfoundry/bosh-init`

To build the micro cli:

- `bin/build` # The `bosh-init` binary will be located in `out/`

Install tools used by the BOSH Micro CLI test suite:

- `bin/go get code.google.com/p/go.tools/cmd/vet`
