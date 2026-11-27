// Copyright 2026 NVIDIA CORPORATION
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

func TestSchedulingShardDefaultsGoMemLimitRatio(t *testing.T) {
	shard := &SchedulingShardSpec{}
	shard.SetDefaultsWhereNeeded()

	assert.Equal(t, 0.9, *shard.GoMemLimitRatio)
}

func TestSchedulingShardConsolidationMinRuntime(t *testing.T) {
	shard := &SchedulingShardSpec{MinRuntime: &MinRuntime{
		PreemptMinRuntime:       ptr.To("1m"),
		ReclaimMinRuntime:       ptr.To("2m"),
		ConsolidationMinRuntime: ptr.To("3m"),
	}}
	shard.SetDefaultsWhereNeeded()
	assert.Equal(t, map[string]string{
		"defaultPreemptMinRuntime":       "1m",
		"defaultReclaimMinRuntime":       "2m",
		"defaultConsolidationMinRuntime": "3m",
	}, shard.Plugins["minruntime"].Arguments)
}
