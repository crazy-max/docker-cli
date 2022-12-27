#!/usr/bin/env bash

set -eu

: "${CLI_DOCS_TOOL_VERSION=v0.5.1}"

export GO111MODULE=auto

function clean {
  rm -rf "$buildir"
}

buildir=$(mktemp -d -t docker-cli-docsgen.XXXXXXXXXX)
trap clean EXIT

(
  set -x
  cp -r . "$buildir/"
  cd "$buildir"
  # init dummy go.mod
  ./scripts/vendor init
  # install cli-docs-tool and copy docs/tools.go in root folder
  # to be able to fetch the required depedencies
  go mod edit -modfile=vendor.mod -require=github.com/docker/cli-docs-tool@${CLI_DOCS_TOOL_VERSION}
  cp docs/tools.go .
  # update vendor
  ./scripts/vendor update
  # build docsgen
  go build -mod=vendor -modfile=vendor.mod -tags docsgen -o /tmp/docsgen ./docs/generate.go
)

(
  set -x
  /tmp/docsgen --formats md --source "$(pwd)/docs/reference/commandline" --target "$(pwd)/docs/reference/commandline"
)

# https://github.com/docker/cli/pull/3924#discussion_r1059986605
mv "$(pwd)/docs/reference/commandline/docker.md" "$(pwd)/docs/reference/commandline/cli.md"

# remove generated help.md file
rm "$(pwd)/docs/reference/commandline/help.md" >/dev/null 2>&1 || true
