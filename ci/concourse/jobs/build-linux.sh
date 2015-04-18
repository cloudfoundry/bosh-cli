#!/usr/bin/env bash

export PATH=/usr/local/ruby/bin:/usr/local/go/bin:$PATH
export GOPATH=$(pwd)/gopath
export GOARCH=amd64
export GOOS=linux

version=`cat version/number`

echo "go is at          -- $(which go)"
echo "working dir is    -- $(pwd)"
echo "working dir has   -- $(ls -la)"
echo "bosh init version -- ${version}"

cd gopath/src/github.com/cloudfoundry/bosh-init
echo $version > VERSION.txt
bin/build
mv bosh-init bosh-init-${version}-linux

# GOARCH=amd64 GOOS=darwin
# GOARCH=amd64 GOOS=windows
