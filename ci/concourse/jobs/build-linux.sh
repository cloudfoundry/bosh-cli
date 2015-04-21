#!/usr/bin/env bash

export PATH=/usr/local/ruby/bin:/usr/local/go/bin:$PATH
export GOPATH=$(pwd)/gopath
export GOARCH=amd64
export GOOS=linux

version=`cat version/number`
filename="bosh-init-${version}-${GOOS}-${GOARCH}"

cd gopath/src/github.com/cloudfoundry/bosh-init

echo "building ${filename}"
cat VERSION.txt

bin/build
ls -la out
mv out/bosh-init out/${filename}
