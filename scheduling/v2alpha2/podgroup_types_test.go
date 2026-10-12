// Copyright 2025 NVIDIA CORPORATION
// SPDX-License-Identifier: Apache-2.0

package v2alpha2

import (
	"testing"
)

func TestParseSafeToDisrupt(t *testing.T) {
	tests := []struct {
		value   string
		want    bool
		wantErr bool
	}{
		{value: "true", want: true},
		{value: "false", want: false},
		{value: "False", want: false},
		{value: "0", want: false},
		{value: "", wantErr: true},
		{value: "no", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := ParseSafeToDisrupt(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseSafeToDisrupt(%q) = %v, want an error", tt.value, *got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseSafeToDisrupt(%q) returned error %v", tt.value, err)
			}
			if got == nil || *got != tt.want {
				t.Fatalf("ParseSafeToDisrupt(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
