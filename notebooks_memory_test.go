package main

import "testing"

func TestParseMemoryGi(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    string
		wantErr bool
	}{
		{name: "whole gibibytes", value: "2", want: "2Gi"},
		{name: "fractional gibibytes", value: "2.0", want: "2Gi"},
		{name: "gpu default", value: "96.0", want: "96Gi"},
		{name: "sub-gibibyte", value: "0.5", want: "512Mi"},
		{name: "invalid", value: "not-a-number", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMemoryGi(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseMemoryGi(%q) expected an error, got %s", tt.value, got.String())
				}
				return
			}
			if err != nil {
				t.Fatalf("parseMemoryGi(%q) unexpected error: %v", tt.value, err)
			}
			if got.String() != tt.want {
				t.Errorf("parseMemoryGi(%q) = %s, want %s", tt.value, got.String(), tt.want)
			}
		})
	}
}
