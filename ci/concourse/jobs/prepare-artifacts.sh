#!/usr/bin/env bash

BOSH_INIT_VERSION=`cat version/number`
echo "working dir is    -- $(pwd)"
echo "working dir has   -- $(ls -la)"
echo "bosh init version -- ${BOSH_INIT_VERSION}"

cd gopath/src/github.com/cloudfoundry/bosh-init
echo $BOSH_INIT_VERSION > version
mv bosh-init bosh-init-${BOSH_INIT_VERSION}-linux

# GOARCH=amd64 GOOS=darwin
# GOARCH=amd64 GOOS=windows
