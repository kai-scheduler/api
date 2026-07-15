# Copyright 2025 NVIDIA CORPORATION
# SPDX-License-Identifier: Apache-2.0

CONTROLLER_TOOLS_VERSION ?= v0.20.1

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

# CRDs derived from Kubernetes projects that require the Kubernetes copyright header.
K8S_COPYRIGHTED_MANIFEST_FILES := config/crd/kai.scheduler_topologies.yaml

.PHONY: all
all: generate manifests clients

.PHONY: build
build: ## Build all packages.
	go build ./...

.PHONY: test
test: ## Run all tests.
	go test ./...

.PHONY: lint
lint: ## Format and vet.
	gofmt -l -w .
	go vet ./...

.PHONY: generate
generate: controller-gen ## Generate DeepCopy method implementations.
	$(CONTROLLER_GEN) object:headerFile="./hack/boilerplate.go.txt" paths="./scheduling/..." paths="./kai/..."

.PHONY: manifests
manifests: controller-gen ## Generate CustomResourceDefinition objects.
	$(CONTROLLER_GEN) crd:allowDangerousTypes=true,generateEmbeddedObjectMeta=true,headerFile="./hack/boilerplate.yaml.txt" paths="./scheduling/..." paths="./kai/..." output:crd:artifacts:config=config/crd

	# Prepend Kubernetes copyright to CRDs derived from Kubernetes projects.
	@for f in $(K8S_COPYRIGHTED_MANIFEST_FILES); do \
		cat ./hack/boilerplate.yaml.kb.txt $$f > $$f.tmp && mv $$f.tmp $$f; \
	done

.PHONY: clients
clients: ## Generate clientset, listers, and informers.
	hack/update-client.sh

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)
