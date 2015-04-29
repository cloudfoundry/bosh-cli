#!/usr/bin/env bash

set -e -x

export PATH=/usr/local/ruby/bin:/usr/local/go/bin:$PATH
export GOPATH=$(pwd)/gopath

base=`pwd`

semver=`cat version-semver/number`
timestamp=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
filename="bosh-init-${semver}-${GOOS}-${GOARCH}"

cd gopath/src/github.com/cloudfoundry/bosh-init

git_rev=`git rev-parse --short HEAD`
version="${semver}-${git_rev}-${timestamp}"

echo "building ${filename} with version ${version}"
sed 's/\[DEV BUILD\]/'"$version"'/' cmd/version.go > cmd/version.tmp && mv cmd/version{.tmp,.go}

bin/build

mkdir $base/out
mv out/bosh-init $base/out/${filename}
