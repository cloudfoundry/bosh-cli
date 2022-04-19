#!/usr/bin/env bash
set -eu -o pipefail

set -x

ensure_not_replace_value() {
  local name=$1
  local value=$(eval echo '$'$name)
  if [ "$value" == 'replace-me' ]; then
    echo "environment variable $name must be set"
    exit 1
  fi
}

set +x
if [[ `whoami` != "root" ]]; then
  echo "acceptance tests must be run as a privileged user"
  exit 1
fi
set -x

export BOSH_INIT_CPI_RELEASE_PATH=$(ls $PWD/cpi-release/*.tgz)
export BOSH_INIT_STEMCELL_PATH=$(ls $PWD/stemcell/*.tgz)

cd bosh-cli

start-garden 1> /dev/null

base=$PWD ./bin/test-acceptance-with-garden
