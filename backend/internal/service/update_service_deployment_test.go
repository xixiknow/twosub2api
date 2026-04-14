//go:build unit

package service

import "testing"

func TestDetectDeploymentMode(t *testing.T) {
	tests := []struct {
		name         string
		buildType    string
		dockerRuntime bool
		want         string
	}{
		{
			name:          "docker runtime overrides release build",
			buildType:     "release",
			dockerRuntime: true,
			want:          "docker",
		},
		{
			name:          "release build outside docker is binary",
			buildType:     "release",
			dockerRuntime: false,
			want:          "binary",
		},
		{
			name:          "source build outside docker stays source",
			buildType:     "source",
			dockerRuntime: false,
			want:          "source",
		},
		{
			name:          "unknown build type falls back to source",
			buildType:     "custom",
			dockerRuntime: false,
			want:          "source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectDeploymentMode(tt.buildType, tt.dockerRuntime); got != tt.want {
				t.Fatalf("detectDeploymentMode(%q, %t) = %q, want %q", tt.buildType, tt.dockerRuntime, got, tt.want)
			}
		})
	}
}
