#!/bin/bash

set -e

cd "$( dirname "${BASH_SOURCE[0]}" )"

exec fly -t "${CONCOURSE_TARGET:-production}" set-pipeline -p bosh-cli:docker-images -c ./pipeline.yml \
  --load-vars-from <(lpass show -G "bosh-cli:docker-images concourse secrets" --notes)
