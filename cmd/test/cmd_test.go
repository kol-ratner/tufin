package cmd_test

import (
	"testing"

	"github.com/kol-ratner/tufin/cmd"
	"github.com/kol-ratner/tufin/internal/config"
)

func TestParseSetFlag(t *testing.T) {
	tests := []struct {
		name     string
		setValue string
		want     map[string][]config.Option
	}{
		{
			name:     "wordpress single option",
			setValue: "wordpress.replicas=2",
			want: map[string][]config.Option{
				"wordpress": {config.WithReplicas(2)},
			},
		},
		{
			name:     "wordpress multiple options",
			setValue: "wordpress.replicas=2,wordpress.memory-request=256Mi",
			want: map[string][]config.Option{
				"wordpress": {
					config.WithReplicas(2),
					config.WithMemoryRequest("256Mi"),
				},
			},
		},
		{
			name:     "multiple components",
			setValue: "wordpress.replicas=2,mysql.replicas=3",
			want: map[string][]config.Option{
				"wordpress": {config.WithReplicas(2)},
				"mysql":     {config.WithReplicas(3)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := cmd.ParseSetFlag(tt.setValue)
			// We'll need to compare the results by applying the options and checking the resulting overrides
			for component, options := range got {
				wantOptions := tt.want[component]
				if len(options) != len(wantOptions) {
					t.Errorf("got %d options for %s, want %d", len(options), component, len(wantOptions))
				}
			}
		})
	}
}
