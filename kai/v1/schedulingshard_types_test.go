// Copyright 2026 NVIDIA CORPORATION
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSchedulingShardDefaultsGoMemLimitRatio(t *testing.T) {
	shard := &SchedulingShardSpec{}
	shard.SetDefaultsWhereNeeded()

	assert.Equal(t, 0.9, *shard.GoMemLimitRatio)
}

func TestScenarioSearchBudgetsDeepCopyCopiesFailedSearchBackoff(t *testing.T) {
	in := &ScenarioSearchBudgets{
		MinFailedSearchBackoff: map[string]metav1.Duration{"reclaim": {Duration: 10 * time.Second}},
		MaxFailedSearchBackoff: map[string]metav1.Duration{"reclaim": {Duration: 2 * time.Minute}},
	}

	out := in.DeepCopy()
	assert.Equal(t, in, out)

	out.MinFailedSearchBackoff["reclaim"] = metav1.Duration{Duration: time.Second}
	out.MaxFailedSearchBackoff["reclaim"] = metav1.Duration{Duration: time.Second}
	assert.Equal(t, 10*time.Second, in.MinFailedSearchBackoff["reclaim"].Duration)
	assert.Equal(t, 2*time.Minute, in.MaxFailedSearchBackoff["reclaim"].Duration)
}
