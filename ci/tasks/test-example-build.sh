#!/usr/bin/env bash
set -eu -o pipefail
set -x

cd bosh-cli

echo "Building docs example..."
go build -o out/example docs
