package deployments_test

import (
	"reflect"
	"testing"

	"github.com/kol-ratner/tufin/internal/config"
	"github.com/kol-ratner/tufin/internal/deployments"
)

func TestShip(t *testing.T) {
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
				"found kubeconfig!",
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
				"found kubeconfig!",
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
				"found kubeconfig!",
				"successfully triggered wordpress deployment",
				"successfully triggered mysql deployment",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msgs := make(chan string, len(tt.wantMsgs))
			err := deployments.Ship(msgs, tt.configs...)

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
