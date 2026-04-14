# KAI Scheduler API

Kubernetes API types and utilities for GPU-aware batch scheduling with KAI Scheduler.

## Overview

This package provides the API definitions, generated clients, and utilities for working with KAI Scheduler's custom resources:

- **Queue** (v2): Hierarchical queue resource for managing workload prioritization
- **PodGroup** (v2alpha2): Batch job grouping with gang scheduling support
- **BindRequest** (v1alpha2): Pod binding coordination for GPU resource allocation

## Installation

```bash
go get github.com/kai-scheduler/api@v0.1.0
```

## Usage

### Working with Queue Resources

```go
import (
    queuev2 "github.com/kai-scheduler/api/api/scheduling/v2"
    "github.com/kai-scheduler/api/client/clientset/versioned"
)

// Create a clientset
config, _ := rest.InClusterConfig()
client, _ := versioned.NewForConfig(config)

// List queues
queues, _ := client.SchedulingV2().Queues().List(context.TODO(), metav1.ListOptions{})
```

### Working with PodGroups

```go
import (
    podgroupv2alpha2 "github.com/kai-scheduler/api/api/scheduling/v2alpha2"
    "github.com/kai-scheduler/api/utilities/podgroup"
)

// Calculate preemptibility from priority
preemptibility := podgroup.CalculatePreemptibility("", int32(50))
// Returns v2alpha2.Preemptible (priority < 100 = preemptible)
```

### GPU Resource Utilities

```go
import (
    "github.com/kai-scheduler/api/utilities/resources"
)

// Check if pod requests GPU fractions
if resources.RequestsGPUFraction(pod) {
    fraction := resources.GetGPUFraction(pod)
    memory := resources.GetGPUMemory(pod)
}

// Check for whole GPU requests
if resources.RequestsWholeGPU(pod) {
    // Pod requests whole GPUs via limits/requests
}

// Extract DRA GPU resources
gpuResources, _ := resources.ExtractDRAGPUResources(ctx, pod, k8sClient)
```

## API Types

### Queue (v2)

Hierarchical queue for workload management with fairshare scheduling:

```go
type Queue struct {
    Spec QueueSpec
    Status QueueStatus
}
```

### PodGroup (v2alpha2)

Gang scheduling primitive for batch workloads:

```go
type PodGroup struct {
    Spec PodGroupSpec
    Status PodGroupStatus
}
```

Features:
- Gang scheduling (min member requirements)
- Hierarchical subgroups with topology awareness
- Preemptibility control (priority-based or explicit)
- Queue assignment

### BindRequest (v1alpha2)

Pod binding request for GPU-aware scheduling:

```go
type BindRequest struct {
    Spec BindRequestSpec
    Status BindRequestStatus
}
```

## CRD Manifests

CRD manifests are included in `config/crd/`:
- `scheduling.run.ai_queues.yaml`
- `scheduling.run.ai_podgroups.yaml`
- `scheduling.run.ai_bindrequests.yaml`

Install with kubectl:
```bash
kubectl apply -f config/crd/
```

## Constants

Common constants for GPU annotations and labels are available in the `constants` package:

```go
import "github.com/kai-scheduler/api/constants"

// GPU fraction annotation
constants.GpuFraction
constants.GpuMemory
constants.GpuPortionContainerName
```

## Contributing

See the [KAI Scheduler repository](https://github.com/kai-scheduler/KAI-Scheduler) for contribution guidelines.

## License

Apache 2.0 License. Copyright 2025 NVIDIA CORPORATION.

## Versioning

This SDK follows [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking API changes (field removals, type changes)
- **MINOR**: New features, new fields (backward compatible)
- **PATCH**: Bug fixes, documentation updates

Current version: **v0.1.0**
