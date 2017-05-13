#!/usr/bin/env bash
set -e
set -x

MICROSERVICE_NAME=armada-stats
TMP_DIR=tmp/build

#workdir to parent directory
cd "$(dirname "${BASH_SOURCE[0]}")/../"

#build armada-stats package
docker build --rm -t "${MICROSERVICE_NAME}" -f Dockerfile.build ./
docker run --rm -v $(pwd):/go/src/github.com/krise3k/armada-stats "${MICROSERVICE_NAME}"
