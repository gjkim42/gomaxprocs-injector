#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

DI_ROOT=$(dirname "${BASH_SOURCE}")/../..

DI_RACE=${DI_RACE-"-race"}

function main {
	testcases=()
	for arg; do
		if [[ "${arg}" == -* ]]; then
			goflags+=("${arg}")
		else
			testcases+=("${arg}")
		fi
	done
	if [[ ${#testcases[@]} -eq 0 ]]; then
		while IFS='' read -r line; do testcases+=("${line}"); done < <(di::test::find_dirs)
	fi
	set -- "${testcases[@]+${testcases[@]}}"

	if [[ -n "${DI_RACE}" ]]; then
		goflags+=("${DI_RACE}")
	fi

	go test "${goflags[@]:+${goflags[@]}}" \
		"${@}"
}

di::test::find_dirs() {
	(
		cd "${DI_ROOT}"
		find -L . -not \( \
			\( \
			  -path './test/e2e/*' \
				-o -path './vendor/*' \
			\) -prune \
		\) -name '*_test.go' -print0 | xargs -0n1 dirname | LC_ALL=C sort -u
	)
}

main "$@"
