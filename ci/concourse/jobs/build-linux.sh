#!/usr/bin/env bash

export PATH=/usr/local/ruby/bin:/usr/local/go/bin:$PATH
export GOPATH=$(pwd)/gopath
export GOARCH=amd64
export GOOS=linux

echo "go is at        -- $(which go)"
echo "working dir is  -- $(pwd)"
echo "working dir has -- $(ls -la)"

cd gopath/src/github.com/cloudfoundry/bosh-init
# TODO: place the version file somewhere
bin/build

# GOARCH=amd64 GOOS=darwin
# GOARCH=amd64 GOOS=windows
