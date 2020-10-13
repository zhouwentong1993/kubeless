#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo /Users/finup123/go/src/k8s.io/code-generator)}

/Users/finup123/go/src/k8s.io/code-generator/generate-groups.sh all \
  redis-trigger/pkg/client redis-trigger/pkg/apis \
  kubeless:v1beta1