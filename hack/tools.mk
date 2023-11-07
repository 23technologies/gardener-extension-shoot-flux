# This make file is supposed to be included in the top-level make file.
# It can be reused by repos vendoring g/g to have some common make recipes for building and installing development
# tools as needed.
# Recipes in the top-level make file should declare dependencies on the respective tool recipes (e.g. $(CONTROLLER_GEN))
# as needed. If the required tool (version) is not built/installed yet, make will make sure to build/install it.
# The *_VERSION variables in this file contain the "default" values, but can be overwritten in the top level make file.

TOOLS_BIN_DIR              := $(TOOLS_DIR)/bin
KO                         := $(TOOLS_BIN_DIR)/ko

# default tool versions
KO_VERSION ?= v0.15.0

#########################################
# Tools                                 #
#########################################

$(KO): $(call tool_version_file,$(KO),$(KO_VERSION))
	GOBIN=$(abspath $(TOOLS_BIN_DIR)) go install github.com/google/ko@$(KO_VERSION)
