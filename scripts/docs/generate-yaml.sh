#!/usr/bin/env bash
# Generate yaml for docker/cli reference docs
set -eu -o pipefail

mkdir -p docs/yaml/gen

GO111MODULE=auto go build -o build/yaml-docs-generator github.com/crazy-max/docker-cli/docs/yaml
build/yaml-docs-generator --root "$(pwd)" --target "$(pwd)/docs/yaml/gen"
