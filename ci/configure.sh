#!/usr/bin/env bash
set -eu -o pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )"

fly -t "${CONCOURSE_TARGET:-bosh}" set-pipeline -p bosh-cli \
    -c ./pipeline.yml
