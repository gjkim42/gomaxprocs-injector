#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

version=${VERSION:-""}
registry=${REGISTRY:-"gjkim42"}
base_image=${BASE_IMAGE:-"gcr.io/distroless/static-debian11"}
go_version=${GO_VERSION:-"1.20"}
os_codename=${OS_CODENAME:-"bullseye"}
output_dir=${OUTPUT_DIR:-"_output"}

docker build \
	-t "${registry}/gomaxprocs-injector:${version}" \
	--build-arg GO_VERSION="${go_version}" \
	--build-arg OS_CODENAME="${os_codename}" \
	--build-arg BASEIMAGE="${base_image}" \
	--build-arg OUTPUT_DIR="${output_dir}" \
	.
