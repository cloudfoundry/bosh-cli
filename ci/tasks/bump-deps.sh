#!/bin/bash

set -e

git clone bosh-cli bumped-bosh-cli

cd bumped-bosh-cli

go get -u ./...
go mod tidy
go mod vendor

if [ "$(git status --porcelain)" != "" ]; then
  git status
  git add vendor go.sum go.mod
  git config user.name "CI Bot"
  git config user.email "cf-bosh-eng@pivotal.io"
  git commit -m "Update vendored dependencies"
fi
