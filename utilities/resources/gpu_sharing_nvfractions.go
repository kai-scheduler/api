// Copyright 2025 NVIDIA CORPORATION
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"strconv"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/kai-scheduler/api/constants"
	schedulingv1alpha2 "github.com/kai-scheduler/api/scheduling/v1alpha2"
)

type nvFractionsAnnotationType int

const (
	nvFractionsRequestAnnotation nvFractionsAnnotationType = iota
	nvFractionsLimitAnnotation
	nvFractionsDevicesAnnotation
	nvFractionsComputeModeAnnotation
	nvFractionsComputeMinGuaranteedAnnotation
	nvFractionsComputeMaxAllowedAnnotation
	nvFractionsComputeAppliedAnnotation
)

func CalcGpuFractionAnnotationForContainer(containerName string) string {
	return constants.NvFractionsAnnotationPrefix + containerName + constants.NvFractionsMemoryRequestSuffix
}

func CalcGpuFractionLimitAnnotationForContainer(containerName string) string {
	return constants.NvFractionsAnnotationPrefix + containerName + constants.NvFractionsMemoryLimitSuffix
}

func CalcGpuVisibleDevicesAnnotationForContainer(containerName string) string {
	return constants.NvFractionsAnnotationPrefix + containerName + constants.NvFractionsVisibleDevicesSuffix
}

func CalcGpuComputeMinGuaranteedAnnotationForContainer(containerName string) string {
	return constants.NvFractionsAnnotationPrefix + containerName + constants.GpuComputeMinGuaranteedSuffix
}

func CalcGpuComputeMaxAllowedAnnotationForContainer(containerName string) string {
	return constants.NvFractionsAnnotationPrefix + containerName + constants.GpuComputeMaxAllowedSuffix
}

func CalcGpuComputeAppliedAnnotationForContainer(containerName string) string {
	return constants.NvFractionsAnnotationPrefix + containerName + constants.GpuComputeAppliedSuffix
}

func ExtractNvFractionsData(pod *v1.Pod) (map[string]NvFractionsContainerRequest, error) {
	fractionsData := make(map[string]NvFractionsContainerRequest)
	for annotationKey, annotationValue := range pod.Annotations {
		if !strings.HasPrefix(annotationKey, constants.NvFractionsAnnotationPrefix) {
			continue
		}
		if isNvFractionsDeviceListAnnotation(annotationKey) {
			continue
		}
		// .gpu-compute.applied is runtime output from gpu-fractioning, not a request.
		if isGpuComputeAppliedAnnotation(annotationKey) {
			continue
		}

		containerName, annotationType, err := parseNvFractionsAnnotationKey(annotationKey)
		if err != nil {
			return nil, err
		}

		containerData := fractionsData[containerName]
		switch annotationType {
		case nvFractionsComputeModeAnnotation:
			if !IsValidGPUComputeSharingMode(annotationValue) {
				return nil, fmt.Errorf("invalid NvFractions compute mode: %s", annotationValue)
			}
			mode := schedulingv1alpha2.GPUComputeSharingMode(annotationValue)
			containerData.ComputeMode = &mode
			fractionsData[containerName] = containerData
			continue
		case nvFractionsComputeMinGuaranteedAnnotation:
			if containerData.ComputePolicy == ComputePolicyMaxAllowed {
				return nil, fmt.Errorf("gpu-compute.min-guaranteed and gpu-compute.max-allowed are mutually exclusive")
			}
			share, err := parseGpuComputeShare(annotationKey, annotationValue)
			if err != nil {
				return nil, err
			}
			containerData.ComputeShare = &share
			containerData.ComputePolicy = ComputePolicyMinGuaranteed
			fractionsData[containerName] = containerData
			continue
		case nvFractionsComputeMaxAllowedAnnotation:
			if containerData.ComputePolicy == ComputePolicyMinGuaranteed {
				return nil, fmt.Errorf("gpu-compute.min-guaranteed and gpu-compute.max-allowed are mutually exclusive")
			}
			share, err := parseGpuComputeShare(annotationKey, annotationValue)
			if err != nil {
				return nil, err
			}
			containerData.ComputeShare = &share
			containerData.ComputePolicy = ComputePolicyMaxAllowed
			fractionsData[containerName] = containerData
			continue
		case nvFractionsRequestAnnotation:
			gpuMemory, err := parseNvFractionsAnnotationValue(annotationKey, annotationValue)
			if err != nil {
				return nil, err
			}
			containerData.Request = &gpuMemory
		case nvFractionsLimitAnnotation:
			gpuMemory, err := parseNvFractionsAnnotationValue(annotationKey, annotationValue)
			if err != nil {
				return nil, err
			}
			containerData.Limit = &gpuMemory
		}

		defaultRequestFromLimit(&containerData)
		fractionsData[containerName] = containerData
	}
	return fractionsData, nil
}

func getNvFractionData(pod *v1.Pod) (*NvFractionsContainerRequest, error) {
	nvFractionsData, err := ExtractNvFractionsData(pod)
	if err != nil {
		return nil, err
	}
	for _, containerData := range nvFractionsData {
		if containerData.Request != nil {
			return &containerData, nil
		}
	}
	return nil, nil
}

// isNvFractionsDeviceListAnnotation reports whether annotationKey is the
// device-list annotation. It shares the NvFractions prefix but, unlike
// request/limit/compute-mode, isn't part of the customer's fractional GPU
// request - it's written by the binder after scheduling - so it must be
// skipped here rather than treated as an invalid key.
func isNvFractionsDeviceListAnnotation(annotationKey string) bool {
	return strings.HasSuffix(annotationKey, constants.NvFractionsVisibleDevicesSuffix)
}

func parseNvFractionsAnnotationKey(annotationKey string) (string, nvFractionsAnnotationType, error) {
	containerNameWithSuffix := strings.TrimPrefix(annotationKey, constants.NvFractionsAnnotationPrefix)
	if strings.HasSuffix(annotationKey, constants.NvFractionsMemoryRequestSuffix) {
		containerName := strings.TrimSuffix(containerNameWithSuffix, constants.NvFractionsMemoryRequestSuffix)
		if containerName == "" {
			return "", 0, fmt.Errorf("invalid NvFractions annotation key: %s", annotationKey)
		}
		return containerName, nvFractionsRequestAnnotation, nil
	}
	if strings.HasSuffix(annotationKey, constants.NvFractionsMemoryLimitSuffix) {
		containerName := strings.TrimSuffix(containerNameWithSuffix, constants.NvFractionsMemoryLimitSuffix)
		if containerName == "" {
			return "", 0, fmt.Errorf("invalid NvFractions annotation key: %s", annotationKey)
		}
		return containerName, nvFractionsLimitAnnotation, nil
	}
	if strings.HasSuffix(annotationKey, constants.NvFractionsVisibleDevicesSuffix) {
		containerName := strings.TrimSuffix(containerNameWithSuffix, constants.NvFractionsVisibleDevicesSuffix)
		if containerName == "" {
			return "", 0, fmt.Errorf("invalid NvFractions annotation key: %s", annotationKey)
		}
		return containerName, nvFractionsDevicesAnnotation, nil
	}
	if strings.HasSuffix(annotationKey, constants.GpuComputeSharingModeSuffix) {
		containerName := strings.TrimSuffix(containerNameWithSuffix, constants.GpuComputeSharingModeSuffix)
		if containerName == "" {
			return "", 0, fmt.Errorf("invalid NvFractions annotation key: %s", annotationKey)
		}
		return containerName, nvFractionsComputeModeAnnotation, nil
	}
	if strings.HasSuffix(annotationKey, constants.GpuComputeMinGuaranteedSuffix) {
		containerName := strings.TrimSuffix(containerNameWithSuffix, constants.GpuComputeMinGuaranteedSuffix)
		if containerName == "" {
			return "", 0, fmt.Errorf("invalid NvFractions annotation key: %s", annotationKey)
		}
		return containerName, nvFractionsComputeMinGuaranteedAnnotation, nil
	}
	if strings.HasSuffix(annotationKey, constants.GpuComputeMaxAllowedSuffix) {
		containerName := strings.TrimSuffix(containerNameWithSuffix, constants.GpuComputeMaxAllowedSuffix)
		if containerName == "" {
			return "", 0, fmt.Errorf("invalid NvFractions annotation key: %s", annotationKey)
		}
		return containerName, nvFractionsComputeMaxAllowedAnnotation, nil
	}
	if strings.HasSuffix(annotationKey, constants.GpuComputeAppliedSuffix) {
		containerName := strings.TrimSuffix(containerNameWithSuffix, constants.GpuComputeAppliedSuffix)
		if containerName == "" {
			return "", 0, fmt.Errorf("invalid NvFractions annotation key: %s", annotationKey)
		}
		return containerName, nvFractionsComputeAppliedAnnotation, nil
	}
	return "", 0, fmt.Errorf("invalid NvFractions annotation key: %s", annotationKey)
}

func parseNvFractionsAnnotationValue(annotationKey, annotationValue string) (resource.Quantity, error) {
	gpuMemory, err := resource.ParseQuantity(annotationValue)
	if err != nil || gpuMemory.Sign() <= 0 {
		return resource.Quantity{}, fmt.Errorf(
			"%s annotation value must be a valid Kubernetes memory quantity greater than 0", annotationKey,
		)
	}
	return gpuMemory, nil
}

func isGpuComputeAppliedAnnotation(annotationKey string) bool {
	return strings.HasSuffix(annotationKey, constants.GpuComputeAppliedSuffix)
}

// parseGpuComputeShare converts a decimal in (0, 1) to 1e-5 units without float rounding.
func parseGpuComputeShare(annotationKey, annotationValue string) (int64, error) {
	raw := strings.TrimSpace(annotationValue)
	if raw == "" {
		return 0, fmt.Errorf("gpu-compute.%s must be a decimal in (0, 1)", policySuffixForError(annotationKey))
	}

	integerPart, fractionPart, hasDecimalPoint := strings.Cut(raw, ".")
	if integerPart == "" || integerPart != "0" {
		return 0, fmt.Errorf("gpu-compute.%s must be a decimal in (0, 1)", policySuffixForError(annotationKey))
	}
	if !hasDecimalPoint || fractionPart == "" {
		return 0, fmt.Errorf("gpu-compute.%s must be a decimal in (0, 1)", policySuffixForError(annotationKey))
	}
	if len(fractionPart) > 5 {
		return 0, fmt.Errorf("gpu-compute.%s must be a decimal in (0, 1)", policySuffixForError(annotationKey))
	}
	for _, r := range fractionPart {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("gpu-compute.%s must be a decimal in (0, 1)", policySuffixForError(annotationKey))
		}
	}
	for _, r := range integerPart {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("gpu-compute.%s must be a decimal in (0, 1)", policySuffixForError(annotationKey))
		}
	}

	padded := fractionPart + strings.Repeat("0", 5-len(fractionPart))
	share, err := strconv.ParseInt(padded, 10, 64)
	if err != nil || share <= 0 || share >= 100000 {
		return 0, fmt.Errorf("gpu-compute.%s must be a decimal in (0, 1)", policySuffixForError(annotationKey))
	}
	return share, nil
}

func policySuffixForError(annotationKey string) string {
	switch {
	case strings.HasSuffix(annotationKey, constants.GpuComputeMinGuaranteedSuffix):
		return string(ComputePolicyMinGuaranteed)
	case strings.HasSuffix(annotationKey, constants.GpuComputeMaxAllowedSuffix):
		return string(ComputePolicyMaxAllowed)
	default:
		return string(ComputePolicyMinGuaranteed)
	}
}

func defaultRequestFromLimit(containerData *NvFractionsContainerRequest) {
	if containerData.Request != nil || containerData.Limit == nil {
		return
	}
	limitCopy := *containerData.Limit
	containerData.Request = &limitCopy
}
