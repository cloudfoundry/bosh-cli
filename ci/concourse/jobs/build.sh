#!/usr/bin/env bash

set -e -x

export PATH=/usr/local/ruby/bin:/usr/local/go/bin:$PATH
export GOPATH=$(pwd)/gopath

semver=`cat version-semver/number`
timestamp=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
filename="bosh-init-${semver}-${GOOS}-${GOARCH}"

cd gopath/src/github.com/cloudfoundry/bosh-init

git_rev=`git rev-parse --short HEAD`
version="${semver}-${git_rev}-${timestamp}"

echo "building ${filename} with version ${version}"
echo $version > VERSION.txt

bin/build
ls -la out
mv out/bosh-init out/${filename}
