#!/bin/bash

set -e

cd "$( dirname "${BASH_SOURCE[0]}" )"

exec fly -t "${CONCOURSE_TARGET:-production}" set-pipeline -p delete-me-bosh-cli -c ./pipeline-test.yml \
  --load-vars-from <(lpass show -G "bosh-cli concourse secrets" --notes)
