#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

find_files() {
	find . -not \( \
		\( \
			-wholename './vendor' \
		\) -prune \
	\) -name '*.go'
}

find_files | xargs gofmt -s -w
find_files | xargs goimports -w

go mod tidy
