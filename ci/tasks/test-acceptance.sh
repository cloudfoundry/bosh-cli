#!/usr/bin/env bash
set -eu -o pipefail

set -x

if [[ $(whoami) != "root" ]]; then
  echo "acceptance tests must be run as a privileged user"
  exit 1
fi

BOSH_INIT_CPI_RELEASE_PATH="$(ls "${PWD}"/cpi-release/*.tgz)"
export BOSH_INIT_CPI_RELEASE_PATH
BOSH_INIT_STEMCELL_PATH="$(ls "${PWD}"/stemcell/*.tgz)"
export BOSH_INIT_STEMCELL_PATH

cd bosh-cli

start-garden 1>/dev/null

bin/test-acceptance-with-garden
