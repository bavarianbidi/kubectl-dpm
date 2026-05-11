// SPDX-License-Identifier: MIT

package profile

import (
	"context"
	"testing"
)

func TestNewBuiltInProfileSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		profileName string
		wantErr     bool
		errContains string
	}{
		{
			name:        "legacy profile",
			profileName: "legacy",
			wantErr:     false,
		},
		{
			name:        "general profile",
			profileName: "general",
			wantErr:     false,
		},
		{
			name:        "baseline profile",
			profileName: "baseline",
			wantErr:     false,
		},
		{
			name:        "restricted profile",
			profileName: "restricted",
			wantErr:     false,
		},
		{
			name:        "netadmin profile",
			profileName: "netadmin",
			wantErr:     false,
		},
		{
			name:        "sysadmin profile",
			profileName: "sysadmin",
			wantErr:     false,
		},
		{
			name:        "invalid profile name",
			profileName: "invalid-profile",
			wantErr:     true,
			errContains: "unknown built-in profile",
		},
		{
			name:        "empty profile name",
			profileName: "",
			wantErr:     true,
			errContains: "unknown built-in profile",
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			source, err := NewBuiltInProfileSource(tt.profileName)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewBuiltInProfileSource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("NewBuiltInProfileSource() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if !tt.wantErr {
				if source == nil {
					t.Error("NewBuiltInProfileSource() returned nil source, want non-nil")
					return
				}

				if got := source.ProfileName(); got != tt.profileName {
					t.Errorf("BuiltInProfileSource.ProfileName() = %v, want %v", got, tt.profileName)
				}
			}
		})
	}
}

func TestBuiltInProfileSource_GetSpec(t *testing.T) {
	t.Parallel()

	source, err := NewBuiltInProfileSource("netadmin")
	if err != nil {
		t.Fatalf("NewBuiltInProfileSource() failed: %v", err)
	}

	spec, err := source.GetSpec(context.Background())
	if err != nil {
		t.Errorf("BuiltInProfileSource.GetSpec() unexpected error = %v", err)
	}

	// Built-in profiles should return nil spec (kubectl handles them)
	if spec != nil {
		t.Errorf("BuiltInProfileSource.GetSpec() = %v, want nil", spec)
	}
}

func TestBuiltInProfileSource_Type(t *testing.T) {
	t.Parallel()

	source, err := NewBuiltInProfileSource("netadmin")
	if err != nil {
		t.Fatalf("NewBuiltInProfileSource() failed: %v", err)
	}

	if got := source.Type(); got != SourceTypeBuiltIn {
		t.Errorf("BuiltInProfileSource.Type() = %v, want %v", got, SourceTypeBuiltIn)
	}
}

func TestBuiltInProfileSource_AllProfiles(t *testing.T) {
	t.Parallel()

	// Test all valid built-in profiles
	profiles := []string{"legacy", "general", "baseline", "restricted", "netadmin", "sysadmin"}

	for _, profileName := range profiles {
		profileName := profileName // Capture range variable
		t.Run(profileName, func(t *testing.T) {
			t.Parallel()

			source, err := NewBuiltInProfileSource(profileName)
			if err != nil {
				t.Errorf("NewBuiltInProfileSource(%q) unexpected error = %v", profileName, err)
				return
			}

			if source.Type() != SourceTypeBuiltIn {
				t.Errorf("Type() = %v, want %v", source.Type(), SourceTypeBuiltIn)
			}

			if source.ProfileName() != profileName {
				t.Errorf("ProfileName() = %v, want %v", source.ProfileName(), profileName)
			}

			spec, err := source.GetSpec(context.Background())
			if err != nil {
				t.Errorf("GetSpec() unexpected error = %v", err)
			}
			if spec != nil {
				t.Errorf("GetSpec() = %v, want nil", spec)
			}
		})
	}
}
