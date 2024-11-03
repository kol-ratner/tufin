package k8s_test

import (
	"path/filepath"
	"testing"

	"github.com/kol-ratner/tufin/pkg/k8s"
	"k8s.io/client-go/rest"
)

func TestGetKubeConfigFromHost(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		wantError  bool
	}{
		{
			name:       "custom kubeconfig path",
			configPath: filepath.Join("testdata", "kubeconfig"),
			wantError:  false,
		},
		{
			name:       "invalid path",
			configPath: "/nonexistent/path",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := k8s.GetKubeConfigFromHost(tt.configPath)
			if (err != nil) != tt.wantError {
				t.Errorf("GetKubeConfigFromHost() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		config    *rest.Config
		wantError bool
	}{
		{
			name: "valid config",
			config: &rest.Config{
				Host: "https://localhost:8443",
				TLSClientConfig: rest.TLSClientConfig{
					Insecure: true,
				},
			},
			wantError: false,
		},
		{
			name: "invalid config",
			config: &rest.Config{
				// Invalid host with no port
				Host: "",
				// Invalid TLS config
				TLSClientConfig: rest.TLSClientConfig{
					CertFile: "/nonexistent/cert",
					KeyFile:  "/nonexistent/key",
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := k8s.NewClient(tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("NewClient() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
