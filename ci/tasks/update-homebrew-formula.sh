#!/usr/bin/env bash

set -e -x -u

version=$(cat version-semver/number)

darwin_cli_sha256=$(shasum -a 256 compiled-darwin/bosh-cli-*-darwin-amd64 | cut -d ' ' -f 1)
linux_cli_sha256=$(shasum -a 256 compiled-linux/bosh-cli-*-linux-amd64 | cut -d ' ' -f 1)

pushd homebrew-tap
  cat <<EOF > bosh-cli.rb
class BoshCli < Formula
  desc "BOSH CLI"
  homepage "https://bosh.io/docs/cli-v2.html"
  version "${version}"

  if OS.mac?
    url "https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-#{version}-darwin-amd64"
    sha256 "${darwin_cli_sha256}"
  elsif OS.linux?
    url "https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-#{version}-linux-amd64"
    sha256 "${linux_cli_sha256}"
  end

  depends_on :arch => :x86_64

  option "with-bosh2", "Rename binary to 'bosh2'. Useful if the old Ruby CLI is needed."

  def install
    binary_name = build.with?("bosh2") ? "bosh2" : "bosh"
    if OS.mac?
      bin.install "bosh-cli-#{version}-darwin-amd64" => binary_name
    elsif OS.linux?
      bin.install "bosh-cli-#{version}-linux-amd64" => binary_name
    end
    (bash_completion/"bosh-cli").write <<-completion
      _#{binary_name}() {
          # All arguments except the first one
          args=("\${COMP_WORDS[@]:1:\$COMP_CWORD}")
          # Only split on newlines
          local IFS=$'\n'
          # Call completion (note that the first element of COMP_WORDS is
          # the executable itself)
          COMPREPLY=(\$(GO_FLAGS_COMPLETION=1 \${COMP_WORDS[0]} "\${args[@]}"))
          return 0
      }
      complete -o default -F _#{binary_name} #{binary_name}
    completion
  end

  test do
    system "#{bin}/#{binary_name} --help"
  end
end
EOF

  git add bosh-cli.rb
  if ! [ -z "$(git status --porcelain)" ];
  then
    git config --global user.email "cf-bosh-eng@pivotal.io"
    git config --global user.name "BOSH CI"
    git commit -m "Release bosh-cli ${version}"
  else
    echo "no new version to commit"
  fi
popd

cp -R homebrew-tap update-brew-formula-output
