#!/usr/bin/env bash

export PATH=/usr/local/ruby/bin:/usr/local/go/bin:$PATH
export GOPATH=$(pwd)/gopath
export GOARCH=amd64
export GOOS=darwin

version=`cat version-semver/number`
filename="bosh-init-${version}-${GOOS}-${GOARCH}"
cat version-label/VERSION.txt

cd gopath/src/github.com/cloudfoundry/bosh-init

echo "building ${filename}"

bin/build
ls -la out
mv out/bosh-init out/${filename}
