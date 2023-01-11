#!/bin/bash
set -eu -o pipefail

cd "$( dirname "${BASH_SOURCE[0]}" )"

exec fly -t bosh-ecosystem set-pipeline -p bosh-cli -c ./pipeline.yml
