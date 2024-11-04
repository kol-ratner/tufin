package deployments_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/kol-ratner/tufin/internal/config"
	"github.com/kol-ratner/tufin/internal/deployments"
	"k8s.io/client-go/kubernetes/fake"
)

func TestShip(t *testing.T) {
	// Create a temporary kubeconfig file for testing
	tmpKubeconfig, err := os.CreateTemp("", "kubeconfig")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpKubeconfig.Name())

	// Write mock kubeconfig content
	mockConfig := `apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://localhost:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token`

	if _, err := tmpKubeconfig.Write([]byte(mockConfig)); err != nil {
		t.Fatal(err)
	}

	// Set KUBECONFIG environment variable
	os.Setenv("KUBECONFIG", tmpKubeconfig.Name())
	defer os.Unsetenv("KUBECONFIG")

	// Rest of your test cases remain the same
	tests := []struct {
		name      string
		configs   []deployments.DeploymentConfig
		wantMsgs  []string
		wantError bool
	}{
		{
			name: "deploy wordpress only",
			configs: []deployments.DeploymentConfig{
				{
					Component: "wordpress",
					Options: []config.Option{
						config.WithReplicas(2),
						config.WithMemoryRequest("256Mi"), // Reduced to be under default limit
					},
				},
			},
			wantMsgs: []string{
				"successfully triggered wordpress deployment",
			},
			wantError: false,
		},
		{
			name: "deploy mysql only",
			configs: []deployments.DeploymentConfig{
				{
					Component: "mysql",
					Options: []config.Option{
						config.WithReplicas(3),
						config.WithCPURequest("500m"),
					},
				},
			},
			wantMsgs: []string{
				"successfully triggered mysql deployment",
			},
			wantError: false,
		},
		{
			name: "deploy both components",
			configs: []deployments.DeploymentConfig{
				{
					Component: "wordpress",
					Options: []config.Option{
						config.WithReplicas(2),
					},
				},
				{
					Component: "mysql",
					Options: []config.Option{
						config.WithReplicas(3),
					},
				},
			},
			wantMsgs: []string{
				"successfully triggered wordpress deployment",
				"successfully triggered mysql deployment",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgs := make(chan string, len(tt.wantMsgs))

			fakeClientset := fake.NewSimpleClientset()
			err := deployments.Ship(msgs, fakeClientset, tt.configs...)

			if (err != nil) != tt.wantError {
				t.Errorf("Ship() error = %v, wantError %v", err, tt.wantError)
			}

			close(msgs)
			gotMsgs := make([]string, 0)
			for msg := range msgs {
				gotMsgs = append(gotMsgs, msg)
			}

			if !reflect.DeepEqual(gotMsgs, tt.wantMsgs) {
				t.Errorf("Ship() messages = %v, want %v", gotMsgs, tt.wantMsgs)
			}
		})
	}
}
