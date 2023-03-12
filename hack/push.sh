#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

version=${VERSION:-""}
registry=${REGISTRY:-"gjkim42"}
release_latest=${RELEASE_LATEST:-"false"}

docker push "${registry}/gomaxprocs-injector:${version}"

if [[ "${release_latest}" == "true" && "${version}" != "latest" ]]; then
	docker tag "${registry}/gomaxprocs-injector:${version}" \
		"${registry}/gomaxprocs-injector:latest"
	docker push "${registry}/gomaxprocs-injector:latest"
fi
