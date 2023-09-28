#!/usr/bin/env bash
set -eu -o pipefail

set -x

concourse_root_dir="$(pwd)"

semver="$(cat version-semver/number)"

filename="${FILENAME_PREFIX}bosh-cli-${semver}-${GOOS}-${GOARCH}"
if [[ $GOOS = 'windows' ]]; then
  filename="${filename}.exe"
fi

cd bosh-cli/

export VERSION_LABEL="${semver}"

bin/build

shasum_value=$(sha1sum out/bosh | cut -f 1 -d' ')
echo "sha1: ${shasum_value}"

mv out/bosh "$concourse_root_dir/compiled-${GOOS}-${GOARCH}/${filename}"
