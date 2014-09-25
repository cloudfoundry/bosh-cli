#!/usr/bin/env bash

set -ex

if [ "$1" == "" ]; then
  echo "At least one argument required. ex: run-in-container.sh /path/to/cmd arg1 arg2"
  exit 1
fi

# Pushing to Docker Hub requires login
DOCKER_IMAGE=${DOCKER_IMAGE:-bosh/micro}

# To push to the Pivotal GoCD Docker Registry (behind firewall):
# DOCKER_IMAGE=docker.gocd.cf-app.com:5000/bosh-container

SRC_DIR=$(cd $(dirname $0)/.. && pwd)
chmod -R o+w $SRC_DIR

echo "Running '$@' in docker container '$DOCKER_IMAGE'..."
docker run \
  -a stderr \
  -v $SRC_DIR:/opt/bosh-micro-cli \
  $DOCKER_IMAGE \
  $@ \
  &

SUBPROC="$!"

trap "
  echo '--------------------- KILLING PROCESS'
  kill $SUBPROC

  echo '--------------------- KILLING CONTAINERS'
  docker ps -q | xargs docker kill
" SIGTERM SIGINT # gocd sends TERM; INT just nicer for testing with Ctrl+C

wait $SUBPROC
