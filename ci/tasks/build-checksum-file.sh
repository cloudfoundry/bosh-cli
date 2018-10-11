#!/usr/bin/env bash
set -eux

base=$PWD

echo "Assets" > binary-checksums
echo -e "\`\`\`" >> binary-checksums
echo "                                                          sha256  file"  >> binary-checksums

pushd compiled-linux
  shasum -a 256 bosh-cli-*-linux-amd64 >> $base/binary-checksums
popd

pushd compiled-darwin
  shasum -a 256 bosh-cli-*-darwin-amd64 >> $base/binary-checksums
popd

pushd compiled-windows
  shasum -a 256 bosh-cli-*-windows-amd64.exe >> $base/binary-checksums
popd
echo -e "\`\`\`" >> binary-checksums

mv binary-checksums checksums/checksums
