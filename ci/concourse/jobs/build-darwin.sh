#!/usr/bin/env bash

export PATH=/usr/local/ruby/bin:/usr/local/go/bin:$PATH
export GOPATH=$(pwd)/gopath
export GOARCH=amd64
export GOOS=darwin

version=`cat version/number`

echo "building bosh-init-${version}-darwin"
echo "- working dir is: $(pwd)"

cd gopath/src/github.com/cloudfoundry/bosh-init
echo $version > VERSION.txt
bin/build
mv out/bosh-init out/bosh-init-${version}-darwin
