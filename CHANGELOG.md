# Changelog

All notable changes to KAI-Scheduler-API will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.0] - 2026-04-13

Initial release of KAI-Scheduler-API as a standalone repository extracted from KAI-Scheduler.

### Added

- **API Types** with full git history preserved:
  - Queue (v2) - Hierarchical queue resource
  - PodGroup (v2alpha2) - Gang scheduling with subgroups
  - BindRequest (v1alpha2) - Pod binding coordination

- **Generated Clients**:
  - Clientset for all API versions
  - Informers for watch/cache patterns
  - Listers for indexed queries

- **CRD Manifests**:
  - `config/crd/scheduling.run.ai_queues.yaml`
  - `config/crd/scheduling.run.ai_podgroups.yaml`
  - `config/crd/scheduling.run.ai_bindrequests.yaml`

- **Utilities**:
  - `utilities/resources/` - GPU sharing utilities:
    - GPU fraction/memory extraction from pod annotations
    - DRA (Dynamic Resource Allocation) support
    - Resource list operations
  - `utilities/podgroup/` - PodGroup business logic:
    - `CalculatePreemptibility()` - Priority-based preemptibility

- **Constants**:
  - GPU annotation keys (GpuFraction, GpuMemory, etc.)
  - Common labels and selectors

### Notes

- Extracted from kai-scheduler v0.14.0 with full commit history (696+ commits)
- Clean separation: Only client-facing API contracts included
- Scheduler-specific utilities (framework handles, webhook wiring) remain in kai-scheduler
- Dependencies: Kubernetes v0.35.3, controller-runtime v0.23.3

[v0.1.0]: https://github.com/kai-scheduler/KAI-Scheduler-API/releases/tag/v0.1.0
