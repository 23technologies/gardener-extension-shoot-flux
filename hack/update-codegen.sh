#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Friendly reminder if workspace location is not in $GOPATH
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
if [ "${SCRIPT_DIR}" != "$(realpath $GOPATH)/src/github.com/stackitcloud/gardener-extension-shoot-flux/hack" ]; then
  cat <<EOF
hack/update-codegen.sh does not work correctly if your workspace is outside GOPATH
because of a know bug in k8s.io/code-generator, see https://github.com/kubernetes/kubernetes/issues/86753.
Please move the workspace to $(realpath $GOPATH)/src/github.com/stackitcloud/gardener-extension-shoot-flux.
EOF
  exit 1
fi

# fetch code-generator module to execute the scripts from the modcache (we don't vendor here)
CODE_GENERATOR_DIR="$(go list -m -tags tools -f '{{ .Dir }}' k8s.io/code-generator)"

rm -f ${GOPATH}/bin/*-gen

# flux API

flux_group() {
  echo "Generating flux API group"

  source "${CODE_GENERATOR_DIR}"/kube_codegen.sh

  kube::codegen::gen_helpers \
      --input-pkg-root github.com/stackitcloud/gardener-extension-shoot-flux/pkg/apis \
      --output-base "${SCRIPT_DIR}/../../../.." \
      --boilerplate "${SCRIPT_DIR}/boilerplate.go.txt"
}

flux_group
