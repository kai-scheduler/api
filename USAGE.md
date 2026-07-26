# Usage

## Working with Queue Resources

```go
import (
    queuev2 "github.com/kai-scheduler/api/scheduling/v2"
    "github.com/kai-scheduler/api/client/clientset/versioned"
)

// Create a clientset
config, _ := rest.InClusterConfig()
client, _ := versioned.NewForConfig(config)

// List queues
queues, _ := client.SchedulingV2().Queues().List(context.TODO(), metav1.ListOptions{})
```

## Working with PodGroups

```go
import (
    podgroupv2alpha2 "github.com/kai-scheduler/api/scheduling/v2alpha2"
    "github.com/kai-scheduler/api/utilities/podgroup"
)

// Calculate preemptibility from priority
preemptibility := podgroup.CalculatePreemptibility("", int32(50))
// Returns v2alpha2.Preemptible (priority < 100 = preemptible)
```

## GPU Resource Utilities

```go
import (
    "github.com/kai-scheduler/api/utilities/resources"
)

// Check if pod requests GPU fractions
if resources.RequestsGPUFraction(pod) {
    fraction := resources.GetGPUFraction(pod)
    memory := resources.GetGPUMemory(pod)
}

// Extract DRA GPU resources
gpuResources, _ := resources.ExtractDRAGPUResources(ctx, pod, k8sClient)
```
