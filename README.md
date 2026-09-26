

# KAI Scheduler API
> [!WARNING]
> 🚧 **Work in progress** 🚧
>
> This module is still under active development. APIs may change without notice and direct production use is not yet recommended.
>
> For now, consume it through KAI Scheduler unless you are prepared to track breaking changes closely.

Kubernetes API types and utilities for GPU-aware batch scheduling with KAI Scheduler.

## Overview

This module provides the API definitions, generated clients, and utilities for working with KAI Scheduler's custom resources:

- **Queue** (`scheduling.run.ai/v2`): Hierarchical queue resource for managing workload prioritization
- **PodGroup** (`scheduling.run.ai/v2alpha2`): Batch job grouping with gang scheduling support
- **BindRequest** (`scheduling.run.ai/v1alpha2`): Pod binding coordination for GPU resource allocation
- **Topology** (`kai.scheduler/v1alpha1`): Topology-aware scheduling input

## Installation

```bash
go get github.com/kai-scheduler/api@v0.1.1
```

## Usage

See [USAGE.md](USAGE.md) for code examples covering Queue resources, PodGroups, and GPU utilities.

## CRD Manifests

CRD manifests are generated into `config/crd/`:
- `scheduling.run.ai_queues.yaml`
- `scheduling.run.ai_podgroups.yaml`
- `scheduling.run.ai_bindrequests.yaml`
- `kai.scheduler_topologies.yaml`

Install with kubectl:
```bash
kubectl apply -f config/crd/
```

## Constants

Common constants for GPU annotations and labels are in the `constants` package:

```go
import "github.com/kai-scheduler/api/constants"

constants.GpuFraction
constants.GpuMemory
```

## Code Generation

Deepcopy methods, CRDs, and clients are generated from the API types:

```bash
make generate   # DeepCopy methods (controller-gen)
make manifests  # CRD manifests into config/crd/
make clients    # clientset, informers, listers (k8s.io/code-generator)
```

## Versioning

Follows [Semantic Versioning](https://semver.org/). `v0.x` until the contract is declared stable.

## License

Apache 2.0 License. Copyright 2025 NVIDIA CORPORATION.

## Contributing

Scheduling and Topology API changes land here first, then kai-scheduler bumps the dependency.
See the [KAI Scheduler repository](https://github.com/kai-scheduler/KAI-Scheduler) for contribution guidelines.
