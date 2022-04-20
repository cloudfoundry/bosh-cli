#!/usr/bin/env bash
set -eu -o pipefail

set -x

cd bosh-cli

bin/clean # TODO: is this needed in CI?
bin/test-unit
