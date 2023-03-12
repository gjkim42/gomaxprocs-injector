VERSION?=$(shell if [ -d .git ]; then git describe --tags --dirty --always; else echo "unknown"; fi)
REGISTRY?=gjkim42
BASEIMAGE=gcr.io/distroless/static-debian11
GO_VERSION?=1.20
OS_CODENAME?=bullseye
OUTPUT_DIR?=_output

.PHONY: build
build:
	go build -o ${OUTPUT_DIR}/gomaxprocs-injector ./cmd/gomaxprocs-injector

.PHONY: update
update:
	hack/update.sh

.PHONY: test
test:
	hack/make-rules/test.sh $(WHAT)

.PHONY: image
image:
	docker build \
		-t ${REGISTRY}/gomaxprocs-injector:${VERSION} \
		--build-arg GO_VERSION=${GO_VERSION} \
		--build-arg OS_CODENAME=${OS_CODENAME} \
		--build-arg BASEIMAGE=${BASEIMAGE} \
		--build-arg OUTPUT_DIR=${OUTPUT_DIR} \
		.

.PHONY: push
push:
	docker push ${REGISTRY}/gomaxprocs-injector:${VERSION}
