#!/usr/bin/env bash
set -eu -o pipefail

set -x

# For ssh tunnel test
/etc/init.d/ssh start

cd bosh-cli

bin/clean # TODO: is this needed in CI?
bin/test-integration
