#!/usr/bin/env bash

set -e -x

export PATH=/usr/local/ruby/bin:/usr/local/go/bin:$PATH
export GOPATH=$(pwd)/gopath

cd gopath/src/github.com/cloudfoundry/bosh-utils
bin/install-ginkgo
bin/test-unit
