#!/bin/bash

set -e

cd "$( dirname "${BASH_SOURCE[0]}" )"

exec fly -t "${CONCOURSE_TARGET:-production}" set-pipeline -p bosh:cli -c ./pipeline.yml \
  --load-vars-from <(lpass show -G "bosh-cli concourse secrets" --notes)
