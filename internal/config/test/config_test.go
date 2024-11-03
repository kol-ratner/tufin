package config_test

import (
	"testing"

	"github.com/kol-ratner/tufin/internal/config"
)

func TestDeploymentOverrides(t *testing.T) {
	tests := []struct {
		name     string
		options  []config.Option
		expected config.DeploymentOverrides
	}{
		{
			name: "single replica override",
			options: []config.Option{
				config.WithReplicas(3),
			},
			expected: config.DeploymentOverrides{
				Replicas: 3,
			},
		},
		{
			name: "resource requests",
			options: []config.Option{
				config.WithCPURequest("500m"),
				config.WithMemoryRequest("1Gi"),
			},
			expected: config.DeploymentOverrides{
				CPURequest:    "500m",
				MemoryRequest: "1Gi",
			},
		},
		{
			name: "resource limits",
			options: []config.Option{
				config.WithCPULimit("1"),
				config.WithMemoryLimit("2Gi"),
			},
			expected: config.DeploymentOverrides{
				CPULimit:    "1",
				MemoryLimit: "2Gi",
			},
		},
		{
			name: "volume size",
			options: []config.Option{
				config.WithVolumeSize("10Gi"),
			},
			expected: config.DeploymentOverrides{
				VolumeSize: "10Gi",
			},
		},
		{
			name: "complete configuration",
			options: []config.Option{
				config.WithReplicas(5),
				config.WithCPURequest("250m"),
				config.WithMemoryRequest("512Mi"),
				config.WithCPULimit("1"),
				config.WithMemoryLimit("2Gi"),
				config.WithVolumeSize("10Gi"),
			},
			expected: config.DeploymentOverrides{
				Replicas:      5,
				CPURequest:    "250m",
				MemoryRequest: "512Mi",
				CPULimit:      "1",
				MemoryLimit:   "2Gi",
				VolumeSize:    "10Gi",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overrides := &config.DeploymentOverrides{}
			for _, opt := range tt.options {
				opt(overrides)
			}

			if overrides.Replicas != tt.expected.Replicas {
				t.Errorf("Replicas = %v, want %v", overrides.Replicas, tt.expected.Replicas)
			}
			if overrides.CPURequest != tt.expected.CPURequest {
				t.Errorf("CPURequest = %v, want %v", overrides.CPURequest, tt.expected.CPURequest)
			}
			if overrides.MemoryRequest != tt.expected.MemoryRequest {
				t.Errorf("MemoryRequest = %v, want %v", overrides.MemoryRequest, tt.expected.MemoryRequest)
			}
			if overrides.CPULimit != tt.expected.CPULimit {
				t.Errorf("CPULimit = %v, want %v", overrides.CPULimit, tt.expected.CPULimit)
			}
			if overrides.MemoryLimit != tt.expected.MemoryLimit {
				t.Errorf("MemoryLimit = %v, want %v", overrides.MemoryLimit, tt.expected.MemoryLimit)
			}
			if overrides.VolumeSize != tt.expected.VolumeSize {
				t.Errorf("VolumeSize = %v, want %v", overrides.VolumeSize, tt.expected.VolumeSize)
			}
		})
	}
}
