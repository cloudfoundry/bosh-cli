#!/usr/bin/env bash
set -eu -o pipefail
set -x

cd bosh-cli

bin/test-unit
