#!/usr/bin/env bash

set -e -x

echo "$DOCKER_IMAGE_TAG" | tee "$(pwd)"/docker-files/tag
