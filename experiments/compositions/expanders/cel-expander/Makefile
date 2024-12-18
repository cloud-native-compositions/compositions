EXPANDER_NAME=cel
GCP_PROJECT_ID ?= $(shell gcloud config get-value project)

# Image URLs to use for building/pushing image targets
GIT_IMG_VERSION ?= $(shell git rev-parse --short HEAD)
IMG_VERSION ?= v0.0.1
IMG_REGISTRY ?= gcr.io/$(GCP_PROJECT_ID)
EXPANDER_IMG ?= $(IMG_REGISTRY)/expander-$(EXPANDER_NAME):$(IMG_VERSION)
EXPANDER_BINARY ?= $(EXPANDER_NAME)
EXPANDER_SERVICE ?= composition-$(EXPANDER_NAME)-v0-0-1
KIND_CLUSTER ?= kind
GOPREFIX ?= GOWORK=off

.PHONY: all
all: build

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Expander CRD and manifests

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(GOPREFIX) $(CONTROLLER_GEN) crd paths="./api/..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen
	$(GOPREFIX) $(CONTROLLER_GEN) object paths="./api/..."

.PHONY: fmt
fmt: license ## Run go fmt against code.
	$(GOPREFIX) go fmt ./...

.PHONY: license
license:
	GOFLAGS= $(GOPREFIX) go run github.com/google/addlicense@04bfe4ee9ca5764577b029acc6a1957fd1997153 -c "Google LLC" -l apache ./

.PHONY: vet
vet: ## Run go vet against code.
	$(GOPREFIX) go vet ./...

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter & yamllint
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix


.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/expanders && $(KUSTOMIZE) edit set image expander-$(EXPANDER_NAME)=${EXPANDER_IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -
	cd config/expanders && $(KUSTOMIZE) edit set image expander-$(EXPANDER_NAME)=expander-$(EXPANDER_NAME):latest

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -


.PHONY: release-manifests
release-manifests: manifests kustomize
	$(KUSTOMIZE) build config/release -o release/manifest.yaml
	$(MAKE) license

##@ Tooling
## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize-$(KUSTOMIZE_VERSION)
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen-$(CONTROLLER_TOOLS_VERSION)
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)

## Tool Versions
KUSTOMIZE_VERSION ?= v5.3.0
CONTROLLER_TOOLS_VERSION ?= v0.14.0
GOLANGCI_LINT_VERSION ?= v1.54.2

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,${GOLANGCI_LINT_VERSION})


# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary (ideally with version)
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f $(1) ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
GOBIN=$(LOCALBIN) $(GOPREFIX) go install $${package} ;\
mv "$$(echo "$(1)" | sed "s/-$(3)$$//")" $(1) ;\
}
endef


###### ----------- Expander ----------------------------

##@ expander pod

.PHONY: build
build: generate
	$(GOPREFIX) go build -v -o bin/${EXPANDER_BINARY} ./

.PHONY: clean
clean: ## clean binary.
	rm -fr bin/${EXPANDER_BINARY}
	docker rmi ${EXPANDER_IMG} .

# If you wish to build the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: build #build ## Build docker image with the manager.
	docker build -t ${EXPANDER_IMG} -f Dockerfile .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${EXPANDER_IMG}

.PHONY: docker-run
docker-run: docker-build
	docker run -p 8443:8443 --entrypoint /expander ${EXPANDER_IMG}

.PHONY: create-kind
create-kind:
	kind delete clusters ${KIND_CLUSTER} || true
	kind create cluster --name ${KIND_CLUSTER}

.PHONY: release-test-kind-manifests
release-test-kind-manifests: manifests kustomize
	mkdir -p release/test
	$(KUSTOMIZE) build config/crd -o release/test/crds.yaml
	cd config/expanders && $(KUSTOMIZE) edit set image expander-$(EXPANDER_NAME)=${EXPANDER_IMG}
	$(KUSTOMIZE) build config/default -o release/test/kind-operator.yaml
	cd config/expanders && $(KUSTOMIZE) edit set image expander-$(EXPANDER_NAME)=expander-$(EXPANDER_NAME):latest

.PHONY: load-kind-images
load-kind-images:
	kind load docker-image ${EXPANDER_IMG} --name ${KIND_CLUSTER}

.PHONY: apply-test-manifests
apply-test-manifests:
	$(KUBECTL) --context kind-${KIND_CLUSTER} create namespace composition-system || true
	$(KUBECTL) --context kind-${KIND_CLUSTER} apply -f release/test/crds.yaml
	sleep 5 # for CRDs to be registered
	$(KUBECTL) --context kind-${KIND_CLUSTER} apply -f release/test/kind-operator.yaml || true # for expander version
	sleep 5

.PHONY: deploy-kind
deploy-kind: release-test-kind-manifests docker-build
	$(MAKE) load-kind-images
	$(MAKE) apply-test-manifests
	kubectl --context kind-${KIND_CLUSTER} get pods -A

.PHONY: unit-test
unit-test: create-kind deploy-kind
	kubectl patch service -n composition-system ${EXPANDER_SERVICE} -p '{"spec":{"type":"LoadBalancer"}}'
	sleep 30
	nodeip=$$(kubectl get nodes -o json  | jq '.items[0].status.addresses[0].address' | xargs echo );\
	nodeport=$$(kubectl get service -n composition-system ${EXPANDER_SERVICE} -o json | jq ".spec.ports[0].nodePort");\
	echo $$nodeip:$$nodeport; \
	$(GOPREFIX) go test -v --addr=$$nodeip:$$nodeport