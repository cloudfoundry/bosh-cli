#!/usr/bin/env bash

export PATH=/usr/local/ruby/bin:/usr/local/go/bin:$PATH
export GOPATH=$(pwd)/gopath

echo "go is at        -- $(which go)"
echo "working dir is  -- $(pwd)"
echo "working dir has -- $(ls -la)"

cd gopath/src/github.com/cloudfoundry/bosh-init
bin/clean
bin/install-ginkgo
bin/test-integration
