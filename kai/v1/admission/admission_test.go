// Copyright 2025 NVIDIA CORPORATION
// SPDX-License-Identifier: Apache-2.0

package admission

import (
	"context"
	"testing"

	"github.com/kai-scheduler/api/constants"
	"github.com/kai-scheduler/api/kai/v1/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

func TestAdmission(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Admission type suite")
}

var _ = Describe("Admission", func() {
	It("Set Defaults", func(ctx context.Context) {
		Admission := &Admission{}
		var replicaCount int32
		replicaCount = 1
		Admission.SetDefaultsWhereNeeded(&replicaCount, nil, common.GpuSharingModeNonMemoryEnforced)
		Expect(*Admission.Service.Enabled).To(Equal(true))
		Expect(*Admission.Service.Image.Name).To(Equal("admission"))
		Expect(*Admission.Replicas).To(Equal(int32(1)))
		Expect(Admission.GPUPodRuntimeClassName).To(BeNil())
		Expect(*Admission.GPUFractionRuntimeClassName).To(Equal(constants.DefaultRuntimeClassName))
	})
	It("Set Defaults with replica count", func(ctx context.Context) {
		Admission := &Admission{}
		var replicaCount int32
		replicaCount = 3
		Admission.SetDefaultsWhereNeeded(&replicaCount, nil, common.GpuSharingModeNonMemoryEnforced)
		Expect(*Admission.Replicas).To(Equal(int32(3)))
	})

	DescribeTable("GpuSharingMode drives admission GPUSharing default",
		func(mode common.GpuSharingMode, gpuSharing bool) {
			Admission := &Admission{}
			Admission.SetDefaultsWhereNeeded(nil, nil, mode)
			Expect(*Admission.GPUSharing).To(Equal(gpuSharing))
		},
		Entry("NonMemoryEnforced", common.GpuSharingModeNonMemoryEnforced, true),
		Entry("HamiCore", common.GpuSharingModeHamiCore, true),
		Entry("NvFractions", common.GpuSharingModeNvFractions, false),
		Entry("Disabled", common.GpuSharingModeDisabled, false),
	)

	It("explicit GPUSharing wins over mode default", func(ctx context.Context) {
		Admission := &Admission{GPUSharing: ptr.To(false)}
		Admission.SetDefaultsWhereNeeded(nil, nil, common.GpuSharingModeHamiCore)
		Expect(*Admission.GPUSharing).To(BeFalse())
	})
	It("does not configure autoscaling by default", func(ctx context.Context) {
		admission := &Admission{}

		admission.SetDefaultsWhereNeeded(nil, nil, common.GpuSharingModeNonMemoryEnforced)

		Expect(admission.Autoscaling).To(BeNil())
	})

	It("sets autoscaling defaults when configured", func(ctx context.Context) {
		admission := &Admission{
			Autoscaling: &Autoscaling{},
		}

		admission.SetDefaultsWhereNeeded(nil, nil, common.GpuSharingModeNonMemoryEnforced)

		Expect(*admission.Autoscaling.Enabled).To(BeFalse())
		Expect(admission.Autoscaling.MaxReplicas).To(BeNil())
		Expect(admission.Autoscaling.TargetInFlightRequests).To(BeNil())
		Expect(admission.Autoscaling.TargetCPUUtilizationPercentage).To(BeNil())
	})

	It("preserves explicit autoscaling configuration", func(ctx context.Context) {
		admission := &Admission{
			Autoscaling: &Autoscaling{
				Enabled:                        ptr.To(true),
				MaxReplicas:                    ptr.To(int32(5)),
				TargetInFlightRequests:         ptr.To(int32(20)),
				TargetCPUUtilizationPercentage: ptr.To(int32(80)),
			},
		}

		admission.SetDefaultsWhereNeeded(nil, nil, common.GpuSharingModeNonMemoryEnforced)

		Expect(*admission.Autoscaling.Enabled).To(BeTrue())
		Expect(*admission.Autoscaling.MaxReplicas).To(Equal(int32(5)))
		Expect(*admission.Autoscaling.TargetInFlightRequests).To(Equal(int32(20)))
		Expect(*admission.Autoscaling.TargetCPUUtilizationPercentage).To(Equal(int32(80)))
	})
})
