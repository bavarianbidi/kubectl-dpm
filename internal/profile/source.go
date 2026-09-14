// SPDX-License-Identifier: MIT

package profile

import (
	"context"
)

const (
	// SourceTypeFile represents a file-based profile source.
	SourceTypeFile = "file"
	// SourceTypeBuiltIn represents a built-in kubectl profile source.
	SourceTypeBuiltIn = "builtin"
	// SourceTypeGit represents a Git repository profile source.
	SourceTypeGit = "git"
	// SourceTypeConfigMap represents a Kubernetes ConfigMap profile source.
	SourceTypeConfigMap = "configmap"
)

// ProfileSource represents a source for debug profile specifications.
// Different implementations can fetch profile data from files, Git repositories,
// ConfigMaps, or represent built-in kubectl profiles.
//
//nolint:revive // ProfileSource is intentionally named this way for clarity
type ProfileSource interface {
	// GetSpec retrieves the raw JSON specification data for the profile.
	// For custom profiles, this returns JSON bytes representing a partial corev1.PodSpec.
	// For built-in profiles, this returns nil (kubectl handles the spec internally).
	GetSpec(ctx context.Context) ([]byte, error)

	// Type returns the source type identifier (e.g., "file", "git", "configmap", "builtin").
	// Used to determine the profile source type and for logging purposes.
	Type() string
}
