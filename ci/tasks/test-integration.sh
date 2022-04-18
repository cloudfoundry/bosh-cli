#!/usr/bin/env bash
set -eu -o pipefail

set -x

# For ssh tunnel test
/etc/init.d/ssh start

cd bosh-cli
bin/clean
bin/test-integration
