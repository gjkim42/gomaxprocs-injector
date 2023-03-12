VERSION?=$(shell if [ -d .git ]; then git describe --tags --dirty --always; else echo "unknown"; fi)
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
	VERSION=${VERSION} hack/image.sh

.PHONY: push
push:
	VERSION=${VERSION} hack/push.sh
