#!/usr/bin/env bash
set -eux

base=$PWD

cp release-notes/release-notes.md binary-checksums

echo "" >> binary-checksums
echo "Assets" >> binary-checksums
echo -e "\`\`\`" >> binary-checksums
echo "                                                          sha256  file"  >> binary-checksums

pushd compiled-linux-amd64
  shasum -a 256 bosh-cli-*-linux-amd64 >> $base/binary-checksums
popd

pushd compiled-darwin-amd64
  shasum -a 256 bosh-cli-*-darwin-amd64 >> $base/binary-checksums
popd

pushd compiled-darwin-arm64
  shasum -a 256 bosh-cli-*-darwin-arm64 >> $base/binary-checksums
popd

pushd compiled-windows-amd64
  shasum -a 256 bosh-cli-*-windows-amd64.exe >> $base/binary-checksums
popd
echo -e "\`\`\`" >> binary-checksums

mv binary-checksums checksums/checksums
