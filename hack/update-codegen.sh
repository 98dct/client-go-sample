#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
echo ${SCRIPT_ROOT}
# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
"${SCRIPT_ROOT}/hack/kube-codegen.sh" "deepcopy,client,informer,lister" \
  "${SCRIPT_ROOT}"/pkg/custom/client \
  "${SCRIPT_ROOT}"/pkg/custom/apis \
  foo:v1alpha1 \
  --output-base "$(dirname "${BASH_SOURCE[0]}")/.." \
  --go-header-file "${SCRIPT_ROOT}"/hack/boilerplate.go.txt

