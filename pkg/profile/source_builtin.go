// SPDX-License-Identifier: MIT

package profile

import (
	"context"
	"fmt"

	kubectldebug "k8s.io/kubectl/pkg/cmd/debug"
)

// BuiltInProfileSource represents a built-in kubectl debug profile.
// Built-in profiles (netadmin, sysadmin, etc.) are handled directly by kubectl
// and don't require custom JSON specifications.
type BuiltInProfileSource struct {
	profileName string
}

// NewBuiltInProfileSource creates a new built-in profile source.
// Returns an error if the profile name is not a recognized built-in profile.
func NewBuiltInProfileSource(name string) (*BuiltInProfileSource, error) {
	// Validate it's a known built-in profile
	validNames := []string{
		// SA1019: ProfileLegacy is deprecated: legacyProfile is planned to be removed in v1.39
		// nolint:staticcheck
		kubectldebug.ProfileLegacy,
		kubectldebug.ProfileGeneral,
		kubectldebug.ProfileBaseline,
		kubectldebug.ProfileRestricted,
		kubectldebug.ProfileNetadmin,
		kubectldebug.ProfileSysadmin,
	}

	valid := false
	for _, v := range validNames {
		if name == v {
			valid = true
			break
		}
	}

	if !valid {
		return nil, fmt.Errorf("unknown built-in profile: %s (valid profiles: %v)", name, validNames)
	}

	return &BuiltInProfileSource{profileName: name}, nil
}

// GetSpec returns nil for built-in profiles as they don't have custom specs.
// The profile name is passed directly to kubectl via the --profile flag.
func (b *BuiltInProfileSource) GetSpec(_ context.Context) ([]byte, error) {
	// Built-in profiles don't have custom specs - kubectl handles them internally
	return nil, nil
}

// Type returns the source type identifier.
func (b *BuiltInProfileSource) Type() string {
	return SourceTypeBuiltIn
}

// ProfileName returns the built-in profile name (e.g., "netadmin", "sysadmin").
func (b *BuiltInProfileSource) ProfileName() string {
	return b.profileName
}
