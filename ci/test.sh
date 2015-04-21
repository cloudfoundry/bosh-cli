#!/bin/bash
set -ex
bin=$(cd $(dirname $0)/../bin && pwd)

$bin/go get code.google.com/p/go.tools/cmd/vet

exec $bin/test
