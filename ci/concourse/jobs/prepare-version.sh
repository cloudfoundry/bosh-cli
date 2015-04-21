#!/usr/bin/env bash

basepath=$(pwd)

echo "basepath: ${basepath}"

cd gopath/src/github.com/cloudfoundry/bosh-init


semver=`cat ${basepath}/version-semver/number`
timestamp=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
git_rev=`git rev-parse --short HEAD`

version="${semver}-${git_rev}-${timestamp}"

echo "version: ${version}"
mkdir -p ${basepath}/prepare-version
echo $version > ${basepath}/prepare-version/current-label
