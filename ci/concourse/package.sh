#!/usr/bin/env bash

export PATH=/usr/local/ruby/bin:/usr/local/go/bin:$PATH
export GOPATH=$(pwd)/gopath

dir=$(pwd)

cd gopath/src/github.com/cloudfoundry/bosh-init
bin/build

cd out
tar zcf bosh-init-0.0.0.tgz bosh-init
mv bosh-init-0.0.0.tgz $dir
