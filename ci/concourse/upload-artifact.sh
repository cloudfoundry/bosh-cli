#!/usr/bin/env bash

export PATH=/usr/local/ruby/bin:/usr/local/go/bin:$PATH
export GOPATH=$(pwd)/gopath

dir=$(pwd)

cd gopath/src/github.com/cloudfoundry/bosh-init
bin/build

cd out
pwd
echo "out dir contains:"
ls -la
echo "moving bosh-init to ${dir}"
mv bosh-init $dir
