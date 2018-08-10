#!/bin/bash

set -e

cd "$( dirname "${BASH_SOURCE[0]}" )"

exec fly -t "${CONCOURSE_TARGET:-production}" set-pipeline -p bosh:cli:registry-removal -c ./pipeline-registry-removal.yml \
  -l <(lpass show -G "bosh-cli concourse secrets" --notes) \
  -l <(lpass show --notes "bosh aws cpi v2 ci secrets")
