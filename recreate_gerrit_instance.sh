#!/bin/bash

set -ex
source .env
GERRIT_ROOT="${GERRIT_ROOT:-gerrit-root}"
GERRIT_PORT="${GERRIT_PORT:-29418}"
GERRIT_WEB_PORT="${GERRIT_WEB_PORT:-8080}"
docker-compose kill gerrit

# Recreate filesystem bits
[ -d ${GERRIT_ROOT} ] && rm -rf ${GERRIT_ROOT}

mkdir -p ${GERRIT_ROOT}/{index,cache,git}

docker-compose up -d gerrit