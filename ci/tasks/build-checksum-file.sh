#!/usr/bin/env bash
set -eux

echo -e "\`\`\`" > binary-checksums
echo "                                                          sha256  file"  >> binary-checksums
shasum -a 256 compiled-linux/bosh-cli-*-linux-amd64 >> binary-checksums
shasum -a 256 compiled-darwin/bosh-cli-*-darwin-amd64 >> binary-checksums
shasum -a 256 compiled-windows/bosh-cli-*-windows-amd64.exe >> binary-checksums
echo -e "\`\`\`" >> binary-checksums

mv binary-checksums checksums/checksums
