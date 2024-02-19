#!/bin/bash
set -eu -o pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )"

exec fly -t cf-fiwg set-pipeline -p bosh-cli -c ./pipeline.yml
