## BOSH Micro CLI [![Build Status](https://travis-ci.org/cloudfoundry/bosh-micro-cli.svg?branch=master)](https://travis-ci.org/cloudfoundry/bosh-micro-cli)

This is the BOSH Micro command line interface in Golang.

### Set up a workstation for development

Note: This guide assumes a few things:

- You have gcc (or an equivalent)
- You can install packages (brew, apt-get, or equivalent)

Get Golang and its dependencies (Mac example, replace with your package manager of choice):

- `brew update`
- `brew install go`
- `brew install git` (Go needs git for the `go get` command)
- `brew install hg` (Go needs mercurial for the `go get` command)

Clone and set up the BOSH Micro CLI repository:

- `go get -d github.com/cloudfoundry/bosh-micro-cli`
- `cd $GOPATH/src/github.com/cloudfoundry/bosh-micro-cli`

From here on out we assume you're working in `$GOPATH/src/github.com/cloudfoundry/bosh-micro-cli`

Install tools used by the BOSH Micro CLI test suite:

- `bin/go get code.google.com/p/go.tools/cmd/vet`
- `bin/go get github.com/golang/lint/golint`

### Running tests

Each package in the agent has its own unit tests and there are integration tests in the `integration` package.
You can also run all tests with `bin/test`.
