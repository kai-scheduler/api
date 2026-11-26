// Copyright 2026 NVIDIA CORPORATION
// SPDX-License-Identifier: Apache-2.0

package crd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

type printColumn struct {
	name     string
	jsonPath string
	priority int32
}

// TestPrinterColumns pins the kubectl get columns of the generated CRDs, so that a regeneration
// that drops or renames a column fails here instead of in a user's terminal.
func TestPrinterColumns(t *testing.T) {
	age := printColumn{"Age", ".metadata.creationTimestamp", 0}
	testCases := map[string][]printColumn{
		"scheduling.run.ai_podgroups.yaml": {
			{"Queue", ".spec.queue", 0},
			{"MinMember", ".spec.minMember", 0},
			{"PriorityClass", ".spec.priorityClassName", 1},
			{"Preemptibility", ".spec.preemptibility", 1},
			{"Unschedulable", ".status.schedulingConditions[-1:].reasons[*].reason", 0},
			age,
		},
		"scheduling.run.ai_queues.yaml": {
			{"Priority", ".spec.priority", 0},
			{"Parent", ".spec.parentQueue", 0},
			{"Children", ".status.childQueues", 0},
			{"DisplayName", ".spec.displayName", 0},
			{"GPUQuota", ".spec.resources.gpu.quota", 0},
			{"GPUAllocated", `.status.allocated.nvidia\.com/gpu`, 0},
			{"GPUOverQuotaWeight", ".spec.resources.gpu.overQuotaWeight", 0},
		},
		"scheduling.run.ai_bindrequests.yaml": {
			{"Pod", ".spec.podName", 0},
			{"Node", ".spec.selectedNode", 0},
			{"Phase", ".status.phase", 0},
			{"Reason", ".status.reason", 1},
			age,
		},
		"kai.scheduler_topologies.yaml": {
			{"Highest", ".spec.levels[0].nodeLabel", 0},
			{"Lowest", ".spec.levels[-1:].nodeLabel", 0},
			{"Levels", ".spec.levels", 1},
			age,
		},
	}

	for file, expected := range testCases {
		t.Run(file, func(t *testing.T) {
			data, err := os.ReadFile(file)
			require.NoError(t, err)
			var crd apiextensionsv1.CustomResourceDefinition
			require.NoError(t, yaml.Unmarshal(data, &crd))
			require.NotEmpty(t, crd.Spec.Versions)

			for _, version := range crd.Spec.Versions {
				if !version.Storage {
					continue
				}
				var actual []printColumn
				for _, column := range version.AdditionalPrinterColumns {
					actual = append(actual, printColumn{column.Name, column.JSONPath, column.Priority})
				}
				assert.Equal(t, expected, actual, "printer columns of storage version %s", version.Name)
			}
		})
	}
}
