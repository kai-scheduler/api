# Copyright 2025 NVIDIA CORPORATION
# SPDX-License-Identifier: Apache-2.0

CONTROLLER_TOOLS_VERSION ?= v0.20.1
CHANGIE_VERSION ?= v1.25.0

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
CHANGIE ?= $(LOCALBIN)/changie

# CRDs derived from Kubernetes projects that require the Kubernetes copyright header.
K8S_COPYRIGHTED_MANIFEST_FILES := config/crd/kai.scheduler_topologies.yaml

.PHONY: all
all: generate manifests clients

.PHONY: build
build: ## Build all packages.
	go build ./...

.PHONY: test
test: ## Run all tests with coverage.
	mkdir -p coverage
	go test -coverprofile=coverage/coverage.out ./...

.PHONY: validate
validate: ## Check formatting and vet without writing files (CI).
	@if [ -n "$$(gofmt -l .)" ]; then echo "Files need formatting:"; gofmt -l .; exit 1; fi
	go vet ./...

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

.PHONY: changie
changie: $(CHANGIE) ## Download changie locally if necessary.
$(CHANGIE): $(LOCALBIN)
	test -s $(LOCALBIN)/changie || GOBIN=$(LOCALBIN) go install github.com/miniscruff/changie@$(CHANGIE_VERSION)

.PHONY: changelog
changelog: changie ## Add a changelog entry. Agents: make changelog KIND=Fixed BODY="...". Humans: make changelog (interactive).
	@if [ -n "$(KIND)" ] && [ -n "$(BODY)" ]; then \
		kind_lower=$$(echo "$(KIND)" | tr '[:upper:]' '[:lower:]'); \
		ts=$$(date '+%Y%m%d-%H%M%S'); \
		out=".changes/unreleased/$${kind_lower}-$${ts}.yaml"; \
		printf 'kind: %s\nbody: |-\n  %s\n' "$(KIND)" "$(BODY)" > "$${out}"; \
		echo "Created $${out}"; \
	elif [ -n "$(KIND)" ] || [ -n "$(BODY)" ]; then \
		echo "Both KIND and BODY must be set for non-interactive mode"; exit 1; \
	else \
		$(CHANGIE) new; \
	fi

.PHONY: changelog-release
changelog-release: changie ## Fold unreleased fragments into CHANGELOG.md as VERSION and clear them. Usage: make changelog-release VERSION=v0.1.1
	@test -n "$(VERSION)" || { echo "VERSION is required, e.g. make changelog-release VERSION=v0.1.1"; exit 1; }
	CHANGIE=$(CHANGIE) bash hack/changelog-fold.sh $(VERSION)

.PHONY: changelog-preview
changelog-preview: changie ## Preview the next release section without writing anything. Usage: make changelog-preview VERSION=v0.1.1
	@test -n "$(VERSION)" || { echo "VERSION is required, e.g. make changelog-preview VERSION=v0.1.1"; exit 1; }
	$(CHANGIE) batch $(VERSION) --dry-run

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)
