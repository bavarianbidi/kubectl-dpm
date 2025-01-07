// SPDX-License-Identifier: MIT

package config

import (
	"reflect"
	"testing"

	"github.com/bavarianbidi/kubectl-dpm/pkg/profile"
)

func TestGenerateConfig(t *testing.T) {
	tests := []struct {
		name           string
		configFile     string
		expectedConfig profile.CustomDebugProfile
		wantErr        bool
	}{
		{
			name:       "valid config file",
			configFile: "test_data/test_config.yaml",
			expectedConfig: profile.CustomDebugProfile{
				Profiles: []profile.Profile{
					{
						ProfileName:     "profile1",
						Profile:         "test_data/profile1.json",
						Image:           "busybox",
						Namespace:       "default",
						ImagePullPolicy: "Always",
					},
					{
						ProfileName: "profile2",
						Profile:     "test_data/profile2.json",
						Image:       "busybox",
					},
					{
						ProfileName: "profile3",
						Profile:     "netadmin",
					},
					{
						ProfileName: "profile4",
						Profile:     "netadmin",
						Image:       "nicolaka/netshoot:v0.13",
						Namespace:   "application",
						MatchLabels: map[string]string{
							"app.kubernetes.io/instance": "app",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "invalid config file",
			configFile: "test_data/invalid_config.yaml",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				// Reset the ConfigurationFile variable
				ConfigurationFile = ""
				// Reset the Config variable
				profile.Config = profile.CustomDebugProfile{}
			})

			// Set the ConfigurationFile variable
			ConfigurationFile = tt.configFile

			// Call GenerateConfig
			err := GenerateConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check the loaded profiles if no error is expected
			if !tt.wantErr {
				if !reflect.DeepEqual(profile.Config, tt.expectedConfig) {
					t.Errorf("expected profiles: %v, got: %v", tt.expectedConfig, profile.Config.Profiles)
				}
			}
		})
	}
}
