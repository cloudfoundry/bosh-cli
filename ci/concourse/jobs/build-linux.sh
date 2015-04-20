#!/usr/bin/env bash

export PATH=/usr/local/ruby/bin:/usr/local/go/bin:$PATH
export GOPATH=$(pwd)/gopath
export GOARCH=amd64
export GOOS=linux

version=`cat version/number`

echo "building bosh-init-${version}-linux"
echo "- working dir is: $(pwd)"

cd gopath/src/github.com/cloudfoundry/bosh-init
echo $version > VERSION.txt
bin/build
mv out/bosh-init out/bosh-init-${version}-linux
