# api Development Guide

`github.com/kai-scheduler/api` is a standalone Go module providing Kubernetes API types, generated clients, and utilities for GPU-aware batch scheduling with KAI Scheduler.

## Repository Structure

```
api/
├── scheduling/              # scheduling.run.ai types
│   ├── v1alpha2/            # BindRequest, NumaPlacementRequest
│   ├── v2/                  # Queue
│   └── v2alpha2/            # PodGroup (+ validating webhook)
├── kai/
│   └── v1alpha1/            # Topology
├── client/                  # Generated clientset, informers, listers (both groups)
├── config/crd/              # Generated CRD YAML manifests
├── constants/               # Public API constants (GPU annotations, labels, priorities)
├── utilities/               # Client-facing utilities
│   ├── resources/           # GPU sharing, DRA extraction, resource helpers
│   └── podgroup/            # PodGroup business logic (preemptibility)
├── hack/                    # Codegen boilerplate + update-client.sh
├── Makefile
├── go.mod
├── README.md
├── CHANGELOG.md
└── LICENSE
```

## What Belongs in This Repository

**Include (client-facing API contracts):**
- CRD type definitions with a generated client — the `scheduling.run.ai` group and `kai.scheduler/v1alpha1` Topology
- Generated Kubernetes clients (clientset, informers, listers)
- CRD manifests (YAML)
- Public-API constants (annotations, labels, priorities)
- Utilities for working with API types (GPU request inspection, DRA support, preemptibility)
- The PodGroup validating webhook (ships with its `v2alpha2` type)

**Exclude (stays in kai-scheduler):**
- `kai.scheduler/v1` config types (Config, SchedulingShard) — operator-internal, controller-runtime only
- Scheduler-framework glue (`k8s_utils`) — pulls `k8s.io/kubernetes`; imports feature gates (would cycle)
- Feature gates and flags (runtime behavior, not API contract)
- Scheduler plugins and infrastructure

## Building and Testing

```bash
make build   # go build ./...
make test    # go test ./...
make lint    # gofmt + go vet
```

## Code Generation

```bash
make generate   # DeepCopy methods (controller-gen object)
make manifests  # CRD manifests into config/crd/ (controller-gen crd)
make clients    # clientset, informers, listers (k8s.io/code-generator)
```

Regenerate when API types change or after bumping Kubernetes dependencies. `make manifests` prepends the
Kubernetes copyright header to `kai.scheduler_topologies.yaml` (derived from Kubernetes projects).

## Versioning & Release

Semantic Versioning; `v0.x` until the contract is declared stable.

1. Update API types/utilities
2. Regenerate (`make generate manifests clients`)
3. Update `CHANGELOG.md`
4. `make build && make test`
5. Commit, tag `vX.Y.Z`, push the tag

## Local Development with kai-scheduler

```go
// kai-scheduler/go.mod (local only — not committed to release branches)
replace github.com/kai-scheduler/api => ../api
```

Make API changes here first, test against kai-scheduler via the `replace` directive, then release and bump
kai-scheduler with `go get github.com/kai-scheduler/api@vX.Y.Z`.

## Code Style

Follow kai-scheduler conventions: three import groups (stdlib, external, internal); snake_case files;
PascalCase exported / camelCase unexported; `Is`/`Has`/`Should` boolean prefixes; Apache 2.0 + NVIDIA
copyright headers on all files; GoDoc on exported symbols; explain "why" not "what".
