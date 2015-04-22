#!/usr/bin/env bash

set -e -x

export PATH=/usr/local/ruby/bin:/usr/local/go/bin:$PATH
export GOPATH=$(pwd)/gopath

cd gopath/src/github.com/cloudfoundry/bosh-init

exec bin/test-acceptance-with-vm-concourse
