// SPDX-License-Identifier: MIT

package profile

import (
	"os"
	"reflect"
	"testing"
)

func TestValidateKubectlPath(t *testing.T) {
	tests := []struct {
		name            string
		wantErr         bool
		config          CustomDebugProfile
		environmentVars map[string]string
	}{
		{
			name: "kubectl path is set",
			config: CustomDebugProfile{
				KubectlPath: "/bin/sh",
			},
		},
		{
			name:   "kubectl from _ var",
			config: CustomDebugProfile{},
			environmentVars: map[string]string{
				"_": "/usr/local/bin/kubectl",
			},
		},
		{
			name: "invalid kubectl - use kubectl from config",
			config: CustomDebugProfile{
				KubectlPath: "/bin/sh",
			},
			environmentVars: map[string]string{
				"_": "/usr/local/bin/kubectl.py",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Config = tt.config

			for k, v := range tt.environmentVars {
				if err := os.Setenv(k, v); err != nil {
					t.Errorf("error setting environment variable %s: %v", k, err)
				}
			}
			if err := ValidateKubectlPath(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateKubectlPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAllProfiles(t *testing.T) {
	tests := []struct {
		name         string
		wantErr      bool
		config       CustomDebugProfile
		wantedConfig CustomDebugProfile
	}{
		{
			name: "valid profiles with duplicate entry",
			config: CustomDebugProfile{
				Profiles: []Profile{
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
						ProfileName: "profile2",
						Profile:     "test_data/profile2.json",
					},
					{
						ProfileName:    "profile3",
						Profile:        "netadmin",
						builtInProfile: true,
					},
				},
			},
			wantedConfig: CustomDebugProfile{
				Profiles: []Profile{
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
						ProfileName:    "profile3",
						Profile:        "netadmin",
						builtInProfile: true,
					},
				},
			},
		},
		{
			name: "invalid profile with missing profile file",
			config: CustomDebugProfile{
				Profiles: []Profile{
					{
						ProfileName:     "profile1",
						Profile:         "test_data/profile1.json",
						Image:           "busybox",
						Namespace:       "default",
						ImagePullPolicy: "Always",
					},
					{
						ProfileName: "profile2",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid profile with missing profile name",
			config: CustomDebugProfile{
				Profiles: []Profile{
					{
						ProfileName:     "profile1",
						Profile:         "test_data/profile1.json",
						Image:           "busybox",
						Namespace:       "default",
						ImagePullPolicy: "Always",
					},
					{
						Profile: "test_data/profile2.json",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid built in profile with missing profile name",
			config: CustomDebugProfile{
				Profiles: []Profile{
					{
						ProfileName:     "profile1",
						Profile:         "test_data/profile1.json",
						Image:           "busybox",
						Namespace:       "default",
						ImagePullPolicy: "Always",
					},
					{
						Profile:        "netadmin",
						builtInProfile: true,
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// set the global Config variable to the test config
			Config = tt.config

			if err := ValidateAllProfiles(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateAllProfiles() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.wantedConfig, Config) {
				t.Errorf("expected: %v, got: %v", tt.wantedConfig, Config)
			}
		})
	}
}

func TestCompleteProfile(t *testing.T) {
	tests := []struct {
		name         string
		profileName  string
		wantErr      bool
		wantedConfig CustomDebugProfile
	}{
		{
			name: "complete nothing in profile1",
			wantedConfig: CustomDebugProfile{
				Profiles: []Profile{
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
						Namespace:   "kube-system",
					},
					{
						ProfileName: "profile3",
						Profile:     "test_data/profile3.json",
					},
				},
			},
			profileName: "profile1",
		},
		{
			name: "complete imagePullPolicy in profile2",
			wantedConfig: CustomDebugProfile{
				Profiles: []Profile{
					{
						ProfileName:     "profile1",
						Profile:         "test_data/profile1.json",
						Image:           "busybox",
						Namespace:       "default",
						ImagePullPolicy: "Always",
					},
					{
						ProfileName:     "profile2",
						Profile:         "test_data/profile2.json",
						Image:           "busybox",
						Namespace:       "kube-system",
						ImagePullPolicy: "IfNotPresent",
					},
					{
						ProfileName: "profile3",
						Profile:     "test_data/profile3.json",
					},
				},
			},
			profileName: "profile2",
		},
		{
			name: "complete everything in profile3",
			wantedConfig: CustomDebugProfile{
				Profiles: []Profile{
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
						Namespace:   "kube-system",
					},
					{
						ProfileName:     "profile3",
						Profile:         "test_data/profile3.json",
						ImagePullPolicy: "IfNotPresent",
					},
				},
			},
			profileName: "profile3",
		},
	}
	for _, tt := range tests {
		Config = CustomDebugProfile{
			Profiles: []Profile{
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
					Namespace:   "kube-system",
				},
				{
					ProfileName: "profile3",
					Profile:     "test_data/profile3.json",
				},
			},
		}
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateProfile(tt.profileName); (err != nil) != tt.wantErr {
				t.Errorf("ValidateProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := CompleteProfile(tt.profileName); (err != nil) != tt.wantErr {
				t.Errorf("CompleteProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && !reflect.DeepEqual(tt.wantedConfig, Config) {
				t.Errorf("expected: %v, got: %v", tt.wantedConfig, Config)
			}
		})
	}
}
