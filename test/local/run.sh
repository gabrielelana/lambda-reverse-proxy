#!/bin/bash

set -xou pipefail

# assert to be run in root directory
if [[ ! -d $PWD/.git ]]; then
    echo "This script should be run in root directory" 1>&2
    exit 1
fi

# build docker image
docker build -t "lrp:latest" .

cd test/local || {
    echo "Cannot change directory to test/local" 1>&2
    exit 1
}

# start test environment and make sure everything is up and running
docker compose -p test-e2e up --build --wait --wait-timeout 60 || {
    docker compose -p test-e2e down --remove-orphans --volumes
    echo "Cannot start docker compose environment" 1>&2
    exit 1
}
while true; do
    curl -s --fail "http://localhost:41414/ping" > /dev/null 2>&1 && break
done

# run test
k6 run --quiet --out json=test_results.json k6.js

EXIT_CODE=$?

# shut down everything
docker compose -p test-e2e down --remove-orphans --volumes

exit $EXIT_CODE
