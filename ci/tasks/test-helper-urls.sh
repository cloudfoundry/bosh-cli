#!/usr/bin/env bash
set -eu -o pipefail

failed=0
chmod +x alpha-release-bucket-linux-amd64/alpha-bosh-cli-*-linux-amd64

# shellcheck disable=SC2211
anchors=$(alpha-release-bucket-linux-amd64/alpha-bosh-cli-*-linux-amd64 --help | awk 'match($0, /#([a-z\-]+)/){ print substr($0, RSTART, RLENGTH)}')
file=$(mktemp)

wget https://bosh.io/docs/cli-v2/ -O "${file}"

echo "---------https://bosh.io/docs/cli-v2/------------"
echo "--------misconfigured URLs in cli helper---------"
for a in ${anchors}; do
  if ! grep "${a}" "${file}" -q; then
    echo "Failed for cmd url: ${a}"
    failed=1
  fi
done

exit ${failed}
