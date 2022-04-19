#!/usr/bin/env bash
set -eu -o pipefail

set -x

cd bosh-cli
bin/clean
bin/test-prepare
bin/test-unit
