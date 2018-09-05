#!/bin/bash

set -e

git clone bosh-cli bumped-bosh-cli

mkdir -p workspace/src/github.com/cloudfoundry/
ln -s $PWD/bumped-bosh-cli workspace/src/github.com/cloudfoundry/bosh-cli

export GOPATH=$PWD/workspace

cd workspace/src/github.com/cloudfoundry/bosh-cli

dep ensure -v -update

if [ "$(git status --porcelain)" != "" ]; then
  git status
  git add vendor Gopkg.lock
  git config user.name "CI Bot"
  git config user.email "cf-bosh-eng@pivotal.io"
  git commit -m "Update vendored dependencies"
fi
