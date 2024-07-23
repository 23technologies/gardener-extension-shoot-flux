#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

BASE_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
export TERM="xterm-256color"

bold="$(tput bold)"
blue="$(tput setaf 4)"
normal="$(tput sgr0)"

function qualify-gvs() {
  APIS_PKG="$1"
  GROUPS_WITH_VERSIONS="$2"
  join_char=""
  res=""

  for GVs in ${GROUPS_WITH_VERSIONS}; do
    IFS=: read -r G Vs <<<"${GVs}"

    for V in ${Vs//,/ }; do
      res="$res$join_char$APIS_PKG/$G/$V"
      join_char=","
    done
  done

  echo "$res"
}

function qualify-gs() {
  APIS_PKG="$1"
  unset GROUPS
  IFS=' ' read -ra GROUPS <<< "$2"
  join_char=""
  res=""

  for G in "${GROUPS[@]}"; do
    res="$res$join_char$APIS_PKG/$G"
    join_char=","
  done

  echo "$res"
}

# setup virtual GOPATH
source "$GARDENER_HACK_DIR"/vgopath-setup.sh

# We need to explicitly pass GO111MODULE=off to k8s.io/code-generator as it is significantly slower otherwise,
# see https://github.com/kubernetes/code-generator/issues/100.
export GO111MODULE=off

rm -f ${GOPATH}/bin/*-gen

# flux API
echo "Generating flux API group"

echo "${bold}Public types${normal}"

echo "Generating ${blue}deepcopy${normal}"
deepcopy-gen \
  --go-header-file "$BASE_DIR/boilerplate.go.txt" \
  --output-file zz_generated.deepcopy.go \
  "$(qualify-gvs "github.com/stackitcloud/gardener-extension-shoot-flux/pkg/apis" "flux:v1alpha1")"

echo "Generating ${blue}defaulter${normal}"
defaulter-gen \
  --go-header-file "$BASE_DIR/boilerplate.go.txt" \
  --output-file zz_generated.defaults.go \
  "$(qualify-gvs "github.com/stackitcloud/gardener-extension-shoot-flux/pkg/apis" "flux:v1alpha1")"
