#!/usr/bin/env bash

set -e -x

source ~/.bashrc

echo $GO_VERSION | tee $(pwd)/metadata-files/version
echo $GO_SHA | tee $(pwd)/metadata-files/sha
echo $DOCKER_IMAGE_TAG | tee $(pwd)/metadata-files/tag

