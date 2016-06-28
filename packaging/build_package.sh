#!/usr/bin/env bash
set -e
set -x

MICROSERVICE_NAME=armada-stats
TMP_DIR=tmp/build


stop_armada-stats_containers()
{
    #disable previous trap
    trap - EXIT HUP INT QUIT PIPE TERM
    armada stop -a "${MICROSERVICE_NAME}"
}

trap stop_armada-stats_containers EXIT HUP INT QUIT PIPE TERM

#workdir to parent directory
cd "$(dirname "${BASH_SOURCE[0]}")/../"

#cleanup tmp dir
rm -fr "$TMP_DIR"

#build armada-stats package
armada build -d local

CONTAINER_ID=$(armada run "${MICROSERVICE_NAME}" --env dev -d local | grep -oh 'Service is running in container [[:alnum:]]*' | awk '{print $NF}')
sleep 5

PACKAGE_VERSION=$(cat VERSION)

armada ssh "$CONTAINER_ID" go run build.go build package
mkdir -p "$TMP_DIR"

for PACKAGE_FILE in "armada-stats_${PACKAGE_VERSION}_amd64.deb" "armada-stats-${PACKAGE_VERSION}-1.x86_64.rpm"
do
	docker cp "$CONTAINER_ID:/go/src/github.com/krise3k/armada-stats/dist/${PACKAGE_FILE}" "$TMP_DIR"
done
