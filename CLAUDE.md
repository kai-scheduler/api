# KAI-Scheduler-API Development Guide

KAI-Scheduler-API is a standalone Go module providing Kubernetes API types, generated clients, and utilities for GPU-aware batch scheduling with KAI Scheduler.

## Repository Structure

```
KAI-Scheduler-API/
├── api/scheduling/          # API type definitions (Queue, PodGroup, BindRequest)
├── client/                  # Generated clientset, informers, listers
├── config/crd/              # CRD YAML manifests
├── constants/               # Public API constants (GPU annotations, labels)
├── utilities/               # Client-facing utilities
│   ├── resources/           # GPU sharing, DRA extraction, resource helpers
│   └── podgroup/            # PodGroup business logic (preemptibility)
├── go.mod
├── README.md
├── CHANGELOG.md
└── LICENSE
```

## What Belongs in This Repository

**Include (Client-Facing API Contracts):**
- API type definitions (Queue, PodGroup, BindRequest)
- Generated Kubernetes clients (clientset, informers, listers)
- CRD manifests (YAML definitions)
- Constants that are part of the public API (annotations, labels)
- Utilities for working with API types (GPU request inspection, DRA support)
- Business logic directly referenced in API documentation (preemptibility calculation)

**Exclude (Scheduler Implementation Details):**
- Scheduler plugin implementations
- Framework handles and infrastructure
- Webhook wiring code (operator-specific)
- Feature gates (runtime behavior, not API contracts)
- Scheduler-specific internal utilities

## Building and Testing

### Build
```bash
go build ./...
```

### Run Tests
```bash
go test ./...
```

### Linting
Use the same linter configuration as kai-scheduler:
```bash
golangci-lint run
```

## Code Generation

Generated clients are created from API types using kubernetes code-generator tools.

**When to regenerate:**
- API types are modified (new fields, new versions)
- After bumping Kubernetes dependencies

**How to regenerate:**
```bash
# TODO: Add code generation scripts when SDK CI/CD is set up
```

## Versioning

This SDK follows [Semantic Versioning](https://semver.org/):

- **MAJOR** (v1.x.x → v2.x.x): Breaking API changes
  - Field removals
  - Type changes
  - Renamed resources

- **MINOR** (v1.0.x → v1.1.x): Backward-compatible additions
  - New API fields (with defaults)
  - New utility functions
  - New CRD versions (with conversion)

- **PATCH** (v1.0.0 → v1.0.1): Bug fixes and documentation
  - Bug fixes in utilities
  - Documentation improvements
  - Non-breaking dependency updates

## Release Process

1. Update API types/utilities as needed
2. Update `CHANGELOG.md` with changes
3. Run tests: `go test ./...`
4. Run build: `go build ./...`
5. Commit changes
6. Tag release: `git tag v0.x.y`
7. Push tag: `git push origin v0.x.y`

## Local Development with kai-scheduler

When making changes to both SDK and kai-scheduler simultaneously:

1. **Make SDK changes first** in this repository
2. **Test locally** using replace directive in kai-scheduler:
   ```go
   // kai-scheduler/go.mod
   replace github.com/kai-scheduler/KAI-Scheduler-API => ../KAI-Scheduler-API
   ```
3. **Run kai-scheduler tests** with local SDK
4. **Release SDK** when ready (tag version)
5. **Update kai-scheduler** to use released SDK version:
   ```bash
   cd kai-scheduler
   go get github.com/kai-scheduler/KAI-Scheduler-API@v0.x.y
   # Remove replace directive
   ```

## Code Style

Follow the same conventions as kai-scheduler:

**Imports**: Three groups (stdlib, external, internal)
```go
import (
    "context"
    "fmt"

    v1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

    "github.com/kai-scheduler/KAI-Scheduler-API/constants"
)
```

**Naming**:
- Files: snake_case (`gpu_sharing.go`, `podgroup.go`)
- Types: PascalCase (`Queue`, `PodGroup`)
- Functions: PascalCase exported, camelCase unexported
- Boolean functions: `Is`/`Has`/`Should` prefix

**Comments**:
- Apache 2.0 + NVIDIA copyright headers on all files
- GoDoc-style for exported functions/types
- Avoid obvious comments; explain "why" not "what"

## Contributing

Changes to API types should be made in this repository first, then consumed by kai-scheduler.

For questions or contributions to the main scheduler, see [kai-scheduler repository](https://github.com/kai-scheduler/KAI-Scheduler).

## Known Issues

### Operator API in Generated Clients

Generated clients currently include references to `kai/v1alpha1` (Topology CRD), which is operator-specific and should remain in kai-scheduler. This creates a dependency on the kai-scheduler module.

**Resolution**: Will be addressed when SDK client generation is set up independently. For now, the dependency exists but doesn't affect SDK users who only work with scheduling APIs.
