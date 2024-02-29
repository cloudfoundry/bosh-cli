#!/usr/bin/env bash

set -e -x

echo "$DOCKER_IMAGE_TAG" | tee "$(pwd)"/docker-files/tag

cp bosh-cli-docker-src/ci/docker/* "$(pwd)"/docker-files/
