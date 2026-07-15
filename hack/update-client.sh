#!/usr/bin/env bash
# Copyright 2025 NVIDIA CORPORATION
# SPDX-License-Identifier: Apache-2.0

# kube_codegen installs tools from the module cache, which has no VCS metadata.
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-buildvcs=false"

# Create an import file so k8s.io/code-generator is part of the module graph.
cat <<EOF > generate-dep.go
package main

import (
	_ "k8s.io/code-generator"
)
EOF

go mod tidy
go mod download

SDK_HACK_DIR="$(cd "$(dirname "$(readlink "$0" || echo "$0")")"; pwd)"
REPO_ROOT="${SDK_HACK_DIR}/.."
CODEGEN_PKG=$(go list -m -f '{{.Dir}}' k8s.io/code-generator)
source ${CODEGEN_PKG}/kube_codegen.sh

# Input is the repo root: gen_client discovers the scheduling.run.ai and
# kai.scheduler/v1alpha1 groups; utilities/constants have no group markers.
kube::codegen::gen_client \
  --boilerplate ${SDK_HACK_DIR}/boilerplate.go.kb.txt \
  --with-watch \
  --output-dir ${REPO_ROOT}/client \
  --output-pkg github.com/kai-scheduler/api/client \
  ${REPO_ROOT}

rm -f generate-dep.go && go mod tidy

changed_files=$(git diff --name-only | grep client/ | grep v1alpha2)
${SDK_HACK_DIR}/replace_headers.sh \
  ${SDK_HACK_DIR}/boilerplate.go.txt \
  ${changed_files}
