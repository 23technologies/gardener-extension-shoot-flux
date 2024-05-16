# This make file is supposed to be included in the top-level make file.
# It can be reused by repos vendoring g/g to have some common make recipes for building and installing development
# tools as needed.
# Recipes in the top-level make file should declare dependencies on the respective tool recipes (e.g. $(CONTROLLER_GEN))
# as needed. If the required tool (version) is not built/installed yet, make will make sure to build/install it.
# The *_VERSION variables in this file contain the "default" values, but can be overwritten in the top level make file.

TOOLS_BIN_DIR              := $(TOOLS_DIR)/bin
KO                         := $(TOOLS_BIN_DIR)/ko
DEEPCOPY_GEN               := $(TOOLS_BIN_DIR)/deepcopy-gen
DEFAULTER_GEN              := $(TOOLS_BIN_DIR)/defaulter-gen

# default tool versions
KO_VERSION ?= v0.15.0
CODE_GENERATOR_VERSION ?= v0.29.5

#########################################
# Tools                                 #
#########################################

$(KO): $(call tool_version_file,$(KO),$(KO_VERSION))
	GOBIN=$(abspath $(TOOLS_BIN_DIR)) go install github.com/google/ko@$(KO_VERSION)

$(DEEPCOPY_GEN):$(call tool_version_file,$(DEEPCOPY_GEN),$(CODE_GENERATOR_VERSION))
	GOBIN=$(abspath $(TOOLS_BIN_DIR)) go install k8s.io/code-generator/cmd/deepcopy-gen@$(CODE_GENERATOR_VERSION)

$(DEFAULTER_GEN): $(call tool_version_file,$(DEFAULTER_GEN),$(CODE_GENERATOR_VERSION))
	GOBIN=$(abspath $(TOOLS_BIN_DIR)) go install k8s.io/code-generator/cmd/defaulter-gen@$(CODE_GENERATOR_VERSION)
