## BOSH Micro CLI [![Build Status](https://travis-ci.org/cloudfoundry/bosh-micro-cli.svg?branch=master)](https://travis-ci.org/cloudfoundry/bosh-micro-cli)

This is the BOSH Micro CLI rewritten to support external CPIs.

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

To build the micro cli:

- `bin/build` # The `bosh-micro` binary will be located in `out/`

Install tools used by the BOSH Micro CLI test suite:

- `bin/go get code.google.com/p/go.tools/cmd/vet`
- `bin/go get github.com/golang/lint/golint`

### Running unit tests

Each package in the agent has its own unit tests and there are integration tests in the `integration` package.
You can also run all tests with `bin/test`.

### Running acceptance tests

Vagrant providers supported are:

* virtualbox
* vmware_fusion
* vmware_workstation
* aws

#### Local provider

  When using a local provider, you may choose to download both the
  bosh-warden-cpi-release and stemcell to a local directory to then be copied
  into the VM. To take advantage of this feature, export the following variables
  prior to running the tests (absolute paths are required):

      $ export BOSH_MICRO_CPI_RELEASE=/tmp/bosh-warden-cpi-9.tgz
      $ export BOSH_MICRO_STEMCELL=/tmp/bosh-stemcell-348-warden-boshlite-ubuntu-trusty-go_agent.tgz
      $ ./bin/test-acceptance-with-vm --provider=virtualbox

#### AWS provider

  When using the AWS provider, you will need to provide the following:

      $ export BOSH_MICRO_PRIVATE_KEY=/tmp/id_rsa
      
      # The following variables are required by Bosh Lite
      $ export BOSH_AWS_ACCESS_KEY_ID=access_key
      $ export BOSH_AWS_SECRET_ACCESS_KEY=secret
      $ export BOSH_LITE_KEYPAIR=keypair
      $ export BOSH_LITE_SECURITY_GROUP=sg-1234
      $ export BOSH_LITE_PRIVATE_KEY=$BOSH_MICRO_PRIVATE_KEY
