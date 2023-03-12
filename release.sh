#!/usr/bin/env bash

export GITHUB_TOKEN=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
pre_tag=$(git describe --tags --abbrev=0 HEAD^)
start=$(git log ${pre_tag}..HEAD --pretty=oneline | awk '{print $1}' | tail -n 1)
end=$(git log ${pre_tag}..HEAD --pretty=oneline | awk '{print $1}' | head -n 1)
echo "pre_tag: ${pre_tag}"
echo "start: ${start}"
echo "end: ${end}"
release-notes \
	--dependencies=false \
	--output=/tmp/RELEASE_NOTES.md \
	--required-author= \
	--branch=main \
	--org=gjkim42 \
	--repo=gomaxprocs-injector \
	--start-sha=${start} \
	--end-sha=${end}
