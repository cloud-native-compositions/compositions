# GCP project to use for pushing manifests
GCP_PROJECT_ID ?= $(shell gcloud config get-value project)

# Image URLs to use for building/pushing image targets
GIT_IMG_VERSION ?= $(shell git rev-parse --short HEAD)
IMG_VERSION ?= v0.0.1
IMG_REGISTRY ?= gcr.io/$(GCP_PROJECT_ID)
IMG ?= $(IMG_REGISTRY)/composition:$(IMG_VERSION)
JINJA_IMG ?= $(IMG_REGISTRY)/expander-jinja2:$(IMG_VERSION)
GETTER_IMG ?= $(IMG_REGISTRY)/expander-getter:$(IMG_VERSION)
GOPREFIX ?= GOWORK=off

KIND_CLUSTER ?= kind

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec


##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: all
all: docker-build

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

DOWNSTREAM_FOLDERS = composition expanders/helm-expander expanders/cel-expander

.PHONY: create-kind-cluster
create-kind-cluster:
	kind delete clusters ${KIND_CLUSTER} || true
	kind create cluster --name ${KIND_CLUSTER}

.PHONY: docker-build $(addprefix docker-build-,$(DOWNSTREAM_FOLDERS))
docker-build: $(addprefix docker-build-,$(DOWNSTREAM_FOLDERS))

.PHONY: release-manifests $(addprefix release-manifests-,$(DOWNSTREAM_FOLDERS))
release-manifests: $(addprefix release-manifests-,$(DOWNSTREAM_FOLDERS))

.PHONY: deploy-kind $(addprefix deploy-kind-,$(DOWNSTREAM_FOLDERS))
deploy-kind: create-kind-cluster $(addprefix deploy-kind-,$(DOWNSTREAM_FOLDERS))

define DOWNSTREAM_template
$(1): $$($(1)_OBJS)
	#commands to build it

docker-build-$(1):
	cd $(1); make docker-build

release-manifests-$(1):
	cd $(1); make release-manifests

deploy-kind-$(1):
	cd $(1); make deploy-kind
endef

$(foreach folder,$(DOWNSTREAM_FOLDERS),$(eval $(call DOWNSTREAM_template,$(folder))))
