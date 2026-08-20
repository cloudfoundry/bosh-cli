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

# Stop the daemon on the way out so the task can't hang on a live dockerd.
trap 'service docker stop' EXIT

# Exports DOCKER_HOST / DOCKER_TLS_JSON and creates director_network.
# Must be sourced. Provided by ghcr.io/cloudfoundry/bosh/docker-cpi.
# shellcheck disable=SC1091
source /usr/local/bin/start-docker

cd bosh-cli

bin/test-acceptance-with-docker
