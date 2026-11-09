// Copyright 2025 NVIDIA CORPORATION
// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/kai-scheduler/api/constants"
	schedulingv1alpha2 "github.com/kai-scheduler/api/scheduling/v1alpha2"
)

func TestParseGpuComputeShare(t *testing.T) {
	tests := []struct {
		name          string
		annotationKey string
		value         string
		wantShare     int64
		wantErr       string
	}{
		{name: "0.29 -> 29000", annotationKey: CalcGpuComputeMinGuaranteedAnnotationForContainer("main"), value: "0.29", wantShare: 29000},
		{name: "0.5 -> 50000", annotationKey: CalcGpuComputeMinGuaranteedAnnotationForContainer("main"), value: "0.5", wantShare: 50000},
		{name: "0.00001 -> 1 (5-digit minimum share)", annotationKey: CalcGpuComputeMaxAllowedAnnotationForContainer("main"), value: "0.00001", wantShare: 1},
		{name: "0.99999 -> 99999 (just under 1)", annotationKey: CalcGpuComputeMaxAllowedAnnotationForContainer("main"), value: "0.99999", wantShare: 99999},
		{name: "0.2 -> 20000", annotationKey: CalcGpuComputeMinGuaranteedAnnotationForContainer("main"), value: "0.2", wantShare: 20000},
		{name: "0.12345 -> 12345 (exactly 5 digits)", annotationKey: CalcGpuComputeMinGuaranteedAnnotationForContainer("main"), value: "0.12345", wantShare: 12345},

		{name: "1.0 rejected", annotationKey: CalcGpuComputeMinGuaranteedAnnotationForContainer("main"), value: "1.0", wantErr: "gpu-compute.min-guaranteed must be a decimal in (0, 1)"},
		{name: "0 rejected", annotationKey: CalcGpuComputeMinGuaranteedAnnotationForContainer("main"), value: "0.0", wantErr: "gpu-compute.min-guaranteed must be a decimal in (0, 1)"},
		{name: "0.0 rejected (zero share)", annotationKey: CalcGpuComputeMaxAllowedAnnotationForContainer("main"), value: "0.00000", wantErr: "gpu-compute.max-allowed must be a decimal in (0, 1)"},
		{name: "1.5 rejected", annotationKey: CalcGpuComputeMinGuaranteedAnnotationForContainer("main"), value: "1.5", wantErr: "gpu-compute.min-guaranteed must be a decimal in (0, 1)"},
		{name: "abc rejected", annotationKey: CalcGpuComputeMinGuaranteedAnnotationForContainer("main"), value: "abc", wantErr: "gpu-compute.min-guaranteed must be a decimal in (0, 1)"},
		{name: "more than 5 decimals rejected", annotationKey: CalcGpuComputeMinGuaranteedAnnotationForContainer("main"), value: "0.123456", wantErr: "gpu-compute.min-guaranteed must be a decimal in (0, 1)"},
		{name: "empty value rejected", annotationKey: CalcGpuComputeMinGuaranteedAnnotationForContainer("main"), value: "", wantErr: "gpu-compute.min-guaranteed must be a decimal in (0, 1)"},
		{name: "missing leading 0 rejected", annotationKey: CalcGpuComputeMinGuaranteedAnnotationForContainer("main"), value: ".5", wantErr: "gpu-compute.min-guaranteed must be a decimal in (0, 1)"},
		{name: "no decimal point rejected", annotationKey: CalcGpuComputeMinGuaranteedAnnotationForContainer("main"), value: "0", wantErr: "gpu-compute.min-guaranteed must be a decimal in (0, 1)"},
		{name: "negative rejected", annotationKey: CalcGpuComputeMinGuaranteedAnnotationForContainer("main"), value: "-0.5", wantErr: "gpu-compute.min-guaranteed must be a decimal in (0, 1)"},
		{name: "max-allowed policy surfaces in error", annotationKey: CalcGpuComputeMaxAllowedAnnotationForContainer("main"), value: "1.5", wantErr: "gpu-compute.max-allowed must be a decimal in (0, 1)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGpuComputeShare(tt.annotationKey, tt.value)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("parseGpuComputeShare() err = nil, want containing %q (got share %d)", tt.wantErr, got)
				}
				if !contains(err.Error(), tt.wantErr) {
					t.Fatalf("parseGpuComputeShare() err = %q, want containing %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseGpuComputeShare() unexpected err: %v", err)
			}
			if got != tt.wantShare {
				t.Fatalf("parseGpuComputeShare() = %d, want %d", got, tt.wantShare)
			}
		})
	}
}

func TestExtractNvFractionsData_ComputeShares(t *testing.T) {
	tests := []struct {
		name              string
		annotations       map[string]string
		want              map[string]NvFractionsContainerRequest
		wantErrContaining string
	}{
		{
			name: "min-guaranteed parsed into fixed-point share with policy",
			annotations: map[string]string{
				constants.NvFractionsAnnotationPrefix + "main" + constants.NvFractionsMemoryRequestSuffix: "1Gi",
				CalcGpuComputeMinGuaranteedAnnotationForContainer("main"):                                 "0.29",
			},
			want: map[string]NvFractionsContainerRequest{
				"main": {
					Request:       quantityPtr("1Gi"),
					ComputeShare:  ptr.To(int64(29000)),
					ComputePolicy: ComputePolicyMinGuaranteed,
				},
			},
		},
		{
			name: "max-allowed parsed into fixed-point share with policy",
			annotations: map[string]string{
				constants.NvFractionsAnnotationPrefix + "main" + constants.NvFractionsMemoryRequestSuffix: "1Gi",
				CalcGpuComputeMaxAllowedAnnotationForContainer("main"):                                    "0.3",
			},
			want: map[string]NvFractionsContainerRequest{
				"main": {
					Request:       quantityPtr("1Gi"),
					ComputeShare:  ptr.To(int64(30000)),
					ComputePolicy: ComputePolicyMaxAllowed,
				},
			},
		},
		{
			name: "compute mode and compute share coexist on the same container",
			annotations: map[string]string{
				constants.NvFractionsAnnotationPrefix + "main" + constants.NvFractionsMemoryRequestSuffix: "1Gi",
				CalcGpuComputeSharingModeAnnotationForContainer("main"):                                   "sm-sharing",
				CalcGpuComputeMaxAllowedAnnotationForContainer("main"):                                    "0.3",
			},
			want: map[string]NvFractionsContainerRequest{
				"main": {
					Request:       quantityPtr("1Gi"),
					ComputeMode:   ptr.To(schedulingv1alpha2.GPUComputeSharingModeSMSharing),
					ComputeShare:  ptr.To(int64(30000)),
					ComputePolicy: ComputePolicyMaxAllowed,
				},
			},
		},
		{
			name: "applied annotation is recognized but not stored on the request",
			annotations: map[string]string{
				constants.NvFractionsAnnotationPrefix + "main" + constants.NvFractionsMemoryRequestSuffix: "1Gi",
				CalcGpuComputeMinGuaranteedAnnotationForContainer("main"):                                 "0.2",
				CalcGpuComputeAppliedAnnotationForContainer("main"):                                       "0.25",
			},
			want: map[string]NvFractionsContainerRequest{
				"main": {
					Request:       quantityPtr("1Gi"),
					ComputeShare:  ptr.To(int64(20000)),
					ComputePolicy: ComputePolicyMinGuaranteed,
				},
			},
		},
		{
			name: "invalid compute share value is rejected by the extractor",
			annotations: map[string]string{
				constants.NvFractionsAnnotationPrefix + "main" + constants.NvFractionsMemoryRequestSuffix: "1Gi",
				CalcGpuComputeMinGuaranteedAnnotationForContainer("main"):                                 "1.5",
			},
			wantErrContaining: "gpu-compute.min-guaranteed must be a decimal in (0, 1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractNvFractionsData(podWithAnnotations(tt.annotations))
			assertErrorContains(t, err, tt.wantErrContaining)
			if tt.wantErrContaining != "" {
				return
			}
			assertNvFractionsDataWithCompute(t, got, tt.want)
		})
	}
}

func TestValidateGPUFractionRequest_GpuComputeFractions(t *testing.T) {
	validMemory := map[string]string{
		constants.GpuFraction: "0.5",
	}
	validContainers := []v1.Container{{Name: "main"}}

	tests := []struct {
		name        string
		annotations map[string]string
		containers  []v1.Container
		wantErr     string
	}{
		{
			name: "range: non-decimal value rejected",
			annotations: merge(validMemory, map[string]string{
				CalcGpuComputeMinGuaranteedAnnotationForContainer("main"): "abc",
			}),
			containers: validContainers,
			wantErr:    "gpu-compute.min-guaranteed must be a decimal in (0, 1)",
		},
		{
			name: "range: more than 5 decimals rejected",
			annotations: merge(validMemory, map[string]string{
				CalcGpuComputeMinGuaranteedAnnotationForContainer("main"): "0.123456",
			}),
			containers: validContainers,
			wantErr:    "gpu-compute.min-guaranteed must be a decimal in (0, 1)",
		},
		{
			name: "range: 1.0 rejected",
			annotations: merge(validMemory, map[string]string{
				CalcGpuComputeMinGuaranteedAnnotationForContainer("main"): "1.0",
			}),
			containers: validContainers,
			wantErr:    "gpu-compute.min-guaranteed must be a decimal in (0, 1)",
		},

		{
			name: "exclusive: both policies on the same container rejected",
			annotations: merge(validMemory, map[string]string{
				CalcGpuComputeMinGuaranteedAnnotationForContainer("main"): "0.3",
				CalcGpuComputeMaxAllowedAnnotationForContainer("main"):    "0.5",
			}),
			containers: validContainers,
			wantErr:    "gpu-compute.min-guaranteed and gpu-compute.max-allowed are mutually exclusive",
		},

		{
			name: "needs memory: compute without any memory fraction rejected",
			annotations: map[string]string{
				CalcGpuComputeMinGuaranteedAnnotationForContainer("main"): "0.3",
			},
			containers: validContainers,
			wantErr:    "gpu-compute.min-guaranteed requires a GPU memory fraction (gpu-fraction, gpu-memory or .gpu-memory.request)",
		},
		{
			name: "needs memory: max-allowed without memory fraction rejected",
			annotations: map[string]string{
				CalcGpuComputeMaxAllowedAnnotationForContainer("main"): "0.3",
			},
			containers: validContainers,
			wantErr:    "gpu-compute.max-allowed requires a GPU memory fraction (gpu-fraction, gpu-memory or .gpu-memory.request)",
		},

		{
			name: "mode: max-allowed with explicit time-slicing rejected",
			annotations: merge(validMemory, map[string]string{
				CalcGpuComputeSharingModeAnnotationForContainer("main"): "time-slicing",
				CalcGpuComputeMaxAllowedAnnotationForContainer("main"):  "0.3",
			}),
			containers: validContainers,
			wantErr:    "gpu-compute.max-allowed is supported only with gpu-compute.mode=sm-sharing",
		},
		{
			name: "mode: max-allowed with default mode (no mode annotation) rejected",
			annotations: merge(validMemory, map[string]string{
				CalcGpuComputeMaxAllowedAnnotationForContainer("main"): "0.3",
			}),
			containers: validContainers,
			wantErr:    "gpu-compute.max-allowed is supported only with gpu-compute.mode=sm-sharing",
		},
		{
			name: "mode: max-allowed with sm-sharing accepted",
			annotations: merge(validMemory, map[string]string{
				CalcGpuComputeSharingModeAnnotationForContainer("main"): "sm-sharing",
				CalcGpuComputeMaxAllowedAnnotationForContainer("main"):  "0.3",
			}),
			containers: validContainers,
			wantErr:    "",
		},
		{
			name: "mode: min-guaranteed with default mode accepted",
			annotations: merge(validMemory, map[string]string{
				CalcGpuComputeMinGuaranteedAnnotationForContainer("main"): "0.3",
			}),
			containers: validContainers,
			wantErr:    "",
		},
		{
			name: "mode: min-guaranteed with explicit time-slicing accepted",
			annotations: merge(validMemory, map[string]string{
				CalcGpuComputeSharingModeAnnotationForContainer("main"):   "time-slicing",
				CalcGpuComputeMinGuaranteedAnnotationForContainer("main"): "0.3",
			}),
			containers: validContainers,
			wantErr:    "",
		},

		{
			name: "container: compute key names a missing container rejected",
			annotations: merge(validMemory, map[string]string{
				CalcGpuComputeMinGuaranteedAnnotationForContainer("missing"): "0.3",
			}),
			containers: validContainers,
			wantErr:    "container missing not found in pod spec",
		},
		{
			name: "container: compute key names a container other than the fraction container rejected",
			annotations: merge(validMemory, map[string]string{
				constants.GpuFractionContainerName:                         "main",
				CalcGpuComputeMinGuaranteedAnnotationForContainer("other"): "0.3",
			}),
			containers: []v1.Container{{Name: "main"}, {Name: "other"}},
			wantErr:    "gpu-fraction-container-name annotation value main does not match container name other",
		},
		{
			name: "container: compute key matches the fraction container accepted",
			annotations: merge(validMemory, map[string]string{
				constants.GpuFractionContainerName:                        "main",
				CalcGpuComputeMinGuaranteedAnnotationForContainer("main"): "0.3",
			}),
			containers: validContainers,
			wantErr:    "",
		},

		{
			name: "nvfractions: memory request + min-guaranteed accepted",
			annotations: map[string]string{
				constants.NvFractionsAnnotationPrefix + "main" + constants.NvFractionsMemoryRequestSuffix: "1Gi",
				CalcGpuComputeMinGuaranteedAnnotationForContainer("main"):                                 "0.2",
			},
			containers: validContainers,
			wantErr:    "",
		},

		{
			name:        "no compute annotations: nothing to validate",
			annotations: validMemory,
			containers:  validContainers,
			wantErr:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{Annotations: tt.annotations},
				Spec:       v1.PodSpec{Containers: tt.containers},
			}
			err := ValidateGPUFractionRequest(pod)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				if err == nil {
					t.Fatalf("ValidateGPUFractionRequest() err = nil, want containing %q", tt.wantErr)
				}
				if !contains(err.Error(), tt.wantErr) {
					t.Fatalf("ValidateGPUFractionRequest() err = %q, want containing %q", err.Error(), tt.wantErr)
				}
			}
		})
	}
}

func TestValidateGPUFractionRequest_ComputeFractionContainerResolution(t *testing.T) {
	tests := []struct {
		name          string
		annotations   map[string]string
		wantContainer string
		wantErr       string
	}{
		{
			name: "min-guaranteed on a non-first container with a compute mode",
			annotations: map[string]string{
				CalcGpuComputeSharingModeAnnotationForContainer("other"):   "sm-sharing",
				CalcGpuComputeMinGuaranteedAnnotationForContainer("other"): "0.2",
			},
			wantContainer: "other",
		},
		{
			name: "max-allowed on a non-first container with sm-sharing",
			annotations: map[string]string{
				CalcGpuComputeSharingModeAnnotationForContainer("other"): "sm-sharing",
				CalcGpuComputeMaxAllowedAnnotationForContainer("other"):  "0.3",
			},
			wantContainer: "other",
		},
		{
			name: "min-guaranteed on a non-first container with the default mode",
			annotations: map[string]string{
				CalcGpuComputeMinGuaranteedAnnotationForContainer("other"): "0.2",
			},
			wantContainer: "other",
		},
		{
			name: "compute fraction on an init container",
			annotations: map[string]string{
				CalcGpuComputeSharingModeAnnotationForContainer("init"):   "sm-sharing",
				CalcGpuComputeMinGuaranteedAnnotationForContainer("init"): "0.2",
			},
			wantContainer: "init",
		},
		{
			name: "compute fractions on two containers rejected",
			annotations: map[string]string{
				CalcGpuComputeMinGuaranteedAnnotationForContainer("main"):  "0.2",
				CalcGpuComputeMinGuaranteedAnnotationForContainer("other"): "0.3",
			},
			wantErr: "currently, kai doesn't support multiple containers with fractional GPU requests",
		},
		{
			name: "different compute policies on two containers rejected",
			annotations: map[string]string{
				CalcGpuComputeMinGuaranteedAnnotationForContainer("main"): "0.2",
				CalcGpuComputeMaxAllowedAnnotationForContainer("other"):   "0.3",
				CalcGpuComputeSharingModeAnnotationForContainer("other"):  "sm-sharing",
			},
			wantErr: "currently, kai doesn't support multiple containers with fractional GPU requests",
		},
		{
			name: "compute fraction conflicts with the explicit fraction container",
			annotations: map[string]string{
				constants.GpuFractionContainerName:                         "main",
				CalcGpuComputeMinGuaranteedAnnotationForContainer("other"): "0.2",
			},
			wantErr: "gpu-fraction-container-name annotation value main does not match container name other",
		},
		{
			name: "compute fraction conflicts with the compute mode container",
			annotations: map[string]string{
				CalcGpuComputeMinGuaranteedAnnotationForContainer("main"): "0.2",
				CalcGpuComputeSharingModeAnnotationForContainer("other"):  "sm-sharing",
			},
			wantErr: "currently, kai doesn't support multiple containers with fractional GPU requests",
		},
	}

	for _, memory := range []struct {
		key   string
		value string
	}{
		{constants.GpuFraction, "0.5"},
		{constants.GpuMemory, "1024"},
	} {
		t.Run(memory.key, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					pod := &v1.Pod{
						ObjectMeta: metav1.ObjectMeta{Annotations: merge(tt.annotations, map[string]string{memory.key: memory.value})},
						Spec: v1.PodSpec{
							Containers:     []v1.Container{{Name: "main"}, {Name: "other"}},
							InitContainers: []v1.Container{{Name: "init"}},
						},
					}
					assertErrorContains(t, ValidateGPUFractionRequest(pod), tt.wantErr)
					if tt.wantErr != "" {
						return
					}
					ref, err := GetFractionContainerRef(pod)
					if !assert.NoError(t, err) {
						return
					}
					assert.Equal(t, tt.wantContainer, ref.Container.Name)
				})
			}
		})
	}
}

func TestExtractNvFractionsData_ConflictingComputePolicies(t *testing.T) {
	pod := podWithAnnotations(map[string]string{
		CalcGpuFractionAnnotationForContainer("main"):             "1Gi",
		CalcGpuComputeMinGuaranteedAnnotationForContainer("main"): "0.2",
		CalcGpuComputeMaxAllowedAnnotationForContainer("main"):    "0.3",
	})
	for i := 0; i < 100; i++ {
		data, err := ExtractNvFractionsData(pod)
		assert.EqualError(t, err, "gpu-compute.min-guaranteed and gpu-compute.max-allowed are mutually exclusive")
		assert.Nil(t, data)
	}
}

func TestComputeAppliedDoesNotSelectFractionContainer(t *testing.T) {
	for _, appliedContainer := range []string{"main", "other", "missing"} {
		for _, hasRequest := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/request=%t", appliedContainer, hasRequest), func(t *testing.T) {
				pod := &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{
						CalcGpuComputeAppliedAnnotationForContainer(appliedContainer): "0.25",
					}},
					Spec: v1.PodSpec{Containers: []v1.Container{{Name: "main"}, {Name: "other"}}},
				}
				if hasRequest {
					pod.Annotations[CalcGpuFractionAnnotationForContainer("main")] = "1Gi"
					pod.Annotations[CalcGpuComputeMinGuaranteedAnnotationForContainer("main")] = "0.2"
				}
				name, found, err := GetNvFractionsContainerName(pod.Annotations)
				assert.NoError(t, err)
				assert.Equal(t, hasRequest, found)
				if hasRequest {
					assert.Equal(t, "main", name)
				} else {
					assert.Empty(t, name)
				}
				ref, err := GetFractionContainerRef(pod)
				if assert.NoError(t, err) {
					assert.Equal(t, "main", ref.Container.Name)
				}
				assert.NoError(t, ValidateGPUFractionRequest(pod))
			})
		}
	}
}

func TestGpuComputeConstants(t *testing.T) {
	tests := []struct {
		got  string
		want string
	}{
		{CalcGpuComputeMinGuaranteedAnnotationForContainer("c"), "nvidia.com/container.c.gpu-compute.min-guaranteed"},
		{CalcGpuComputeMaxAllowedAnnotationForContainer("c"), "nvidia.com/container.c.gpu-compute.max-allowed"},
		{CalcGpuComputeAppliedAnnotationForContainer("c"), "nvidia.com/container.c.gpu-compute.applied"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("got %q, want %q", tt.got, tt.want)
		}
	}
}

func TestNvFractionNodeReadyReasonConstants(t *testing.T) {
	assert.Equal(t, "MemoryOnly", constants.NvFractionNodeReadyReasonMemoryOnly)
	assert.Equal(t, "Full", constants.NvFractionNodeReadyReasonFull)
}

// merge returns a copy of a with b's entries taking precedence.
func merge(a, b map[string]string) map[string]string {
	out := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsString(s, substr))
}

func containsString(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func assertNvFractionsDataWithCompute(
	t *testing.T, got map[string]NvFractionsContainerRequest, want map[string]NvFractionsContainerRequest,
) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("ExtractNvFractionsData() returned %d containers, want %d: %#v", len(got), len(want), got)
	}
	for containerName, wantData := range want {
		gotData, ok := got[containerName]
		if !ok {
			t.Fatalf("ExtractNvFractionsData() missing container %q in %#v", containerName, got)
		}
		assertQuantityString(t, gotData.Request, quantityString(wantData.Request))
		assertQuantityString(t, gotData.Limit, quantityString(wantData.Limit))
		if gotData.ComputeMode == nil && wantData.ComputeMode == nil {
		} else if gotData.ComputeMode == nil || wantData.ComputeMode == nil || *gotData.ComputeMode != *wantData.ComputeMode {
			t.Fatalf("ComputeMode = %v, want %v", gotData.ComputeMode, wantData.ComputeMode)
		}
		if (gotData.ComputeShare == nil) != (wantData.ComputeShare == nil) {
			t.Fatalf("ComputeShare = %v, want %v", gotData.ComputeShare, wantData.ComputeShare)
		}
		if gotData.ComputeShare != nil && *gotData.ComputeShare != *wantData.ComputeShare {
			t.Fatalf("ComputeShare = %d, want %d", *gotData.ComputeShare, *wantData.ComputeShare)
		}
		if gotData.ComputePolicy != wantData.ComputePolicy {
			t.Fatalf("ComputePolicy = %q, want %q", gotData.ComputePolicy, wantData.ComputePolicy)
		}
	}
}
