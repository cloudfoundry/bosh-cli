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

/var/vcap/jobs/garden/bin/pre-start
/var/vcap/jobs/garden/bin/garden_ctl start &
/var/vcap/jobs/garden/bin/post-start

bin/test-acceptance-with-garden
