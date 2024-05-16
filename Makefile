# SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

GARDENER_HACK_DIR := $(shell go list -m -f '{{.Dir}}' github.com/gardener/gardener)/hack

EXTENSION_PREFIX            := gardener-extension
NAME                        := shoot-flux
REPO 						:= ghcr.io/stackitcloud/gardener-extension-shoot-flux
REPO_ROOT                   := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
HACK_DIR                    := $(REPO_ROOT)/hack
VERSION                     := $(shell git describe --tag --always --dirty)
TAG							:= $(VERSION)
LD_FLAGS                    := -w $(shell EFFECTIVE_VERSION=$(VERSION) bash $(GARDENER_HACK_DIR)/get-build-ld-flags.sh k8s.io/component-base $(REPO_ROOT)/go.mod $(EXTENSION_PREFIX)-$(NAME) 2>&1 | grep -v .dockerignore)
LEADER_ELECTION             := false
IGNORE_OPERATION_ANNOTATION := false

SHELL=/usr/bin/env bash -o pipefail

#########################################
# Tools                                 #
#########################################

TOOLS_DIR := $(HACK_DIR)/tools
-include $(GARDENER_HACK_DIR)/tools.mk
include hack/tools.mk

.PHONY: start
start:
	@LEADER_ELECTION_NAMESPACE=garden GO111MODULE=on go run \
		-ldflags $(LD_FLAGS) \
		./cmd/$(EXTENSION_PREFIX)-$(NAME) \
		--kubeconfig=${KUBECONFIG} \
		--ignore-operation-annotation=$(IGNORE_OPERATION_ANNOTATION) \
		--leader-election=$(LEADER_ELECTION)

.PHONY: debug
debug:
	@LEADER_ELECTION_NAMESPACE=garden GO111MODULE=on dlv debug\
		./cmd/$(EXTENSION_PREFIX)-$(NAME) -- \
		--kubeconfig=${KUBECONFIG} \
		--ignore-operation-annotation=$(IGNORE_OPERATION_ANNOTATION) \
		--leader-election=$(LEADER_ELECTION)

#################################################################
# Rules related to binary build, Docker image build and release #
#################################################################

PUSH ?= false
images: export KO_DOCKER_REPO = $(REPO)
images: export LD_FLAGS := $(LD_FLAGS)

.PHONY: images
images: $(KO)
	KO_DOCKER_REPO=$(REPO) $(KO) build --sbom none -t $(TAG) --bare --platform linux/amd64,linux/arm64 --push=$(PUSH) ./cmd/gardener-extension-shoot-flux

#####################################################################
# Rules for verification, formatting, linting, testing and cleaning #
#####################################################################

.PHONY: tidy
tidy:
	go mod tidy

# run `make init` to perform an initial go mod cache sync which is required for other make targets
init: tidy

.PHONY: clean
clean:
	@bash $(GARDENER_HACK_DIR)/clean.sh ./cmd/... ./pkg/...

.PHONY: check-generate
check-generate:
	@bash $(GARDENER_HACK_DIR)/check-generate.sh $(REPO_ROOT)

.PHONY: check
check: $(GO_ADD_LICENSE) $(GOIMPORTS) $(GOLANGCI_LINT) $(HELM) $(YQ)
	@bash $(GARDENER_HACK_DIR)/check.sh --golangci-lint-config=./.golangci.yaml ./cmd/... ./pkg/...
	@bash $(GARDENER_HACK_DIR)/check-charts.sh ./charts

.PHONY: generate
generate: $(VGOPATH) $(DEEPCOPY_GEN) $(DEFAULTER_GEN) $(GEN_CRD_API_REFERENCE_DOCS) $(HELM)
	@bash $(GARDENER_HACK_DIR)/generate-controller-registration.sh --pod-security-enforce=privileged shoot-flux charts/gardener-extension-shoot-flux latest deploy/extension/base/controller-registration.yaml Extension:shoot-flux
	@VGOPATH=$(VGOPATH) \
	REPO_ROOT=$(REPO_ROOT) \
	GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) \
	bash $(GARDENER_HACK_DIR)/generate-sequential.sh ./cmd/... ./pkg/...
	@gen-crd-api-reference-docs -api-dir ./pkg/apis/flux/v1alpha1 -config ./hack/api-reference/api.json -template-dir $(GARDENER_HACK_DIR)/api-reference/template -out-file ./hack/api-reference/api.md

.PHONY: format
format: $(GOIMPORTS) $(GOIMPORTSREVISER)
	@bash $(GARDENER_HACK_DIR)/format.sh ./cmd ./pkg

.PHONY: test
test: $(REPORT_COLLECTOR)
	@./hack/test.sh ./cmd/... ./pkg/...

.PHONY: test-cov
test-cov:
	@bash $(GARDENER_HACK_DIR)/test-cover.sh ./cmd/... ./pkg/...

.PHONY: test-cov-clean
test-cov-clean:
	@bash $(GARDENER_HACK_DIR)/test-cover-clean.sh

.PHONY: verify
verify: check format test

.PHONY: verify-extended
verify-extended: check-generate check format test

#####################################################################
# Rules for local environment                                       #
#####################################################################

# speed-up skaffold deployments by building all images concurrently
extension-%: export SKAFFOLD_BUILD_CONCURRENCY = 0
extension-%: export SKAFFOLD_DEFAULT_REPO = localhost:5001
extension-%: export SKAFFOLD_PUSH = true
# use static label for skaffold to prevent rolling all gardener components on every `skaffold` invocation
extension-%: export SKAFFOLD_LABEL = skaffold.dev/run-id=shoot-flux

extension-up: $(SKAFFOLD)
	$(SKAFFOLD) run
extension-dev: $(SKAFFOLD)
	$(SKAFFOLD) dev --cleanup=false --trigger=manual
extension-down: $(SKAFFOLD)
	$(SKAFFOLD) delete
