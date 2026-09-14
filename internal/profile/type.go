// SPDX-License-Identifier: MIT

package profile

import (
	corev1 "k8s.io/api/core/v1"
)

type Profile struct {
	ProfileName     string              `koanf:"name" yaml:"name" validate:"required"`
	Profile         string              `koanf:"profile" yaml:"profile"`             // DEPRECATED: use ProfileSource instead
	ProfileSource   ProfileSourceConfig `koanf:"profileSource" yaml:"profileSource"` // NEW: flexible profile source configuration
	Image           string              `koanf:"image" yaml:"image" validate:"required"`
	Namespace       string              `koanf:"namespace" yaml:"namespace" validate:"required"`
	ImagePullPolicy corev1.PullPolicy   `koanf:"imagePullPolicy" yaml:"imagePullPolicy" validate:"required"`
	TargetContainer string              `koanf:"targetContainer" yaml:"targetContainer" validate:"required"`
	MatchLabels     map[string]string   `koanf:"matchLabels" yaml:"matchLabels" validate:"required"`

	// only used internally
	builtInProfile bool
	source         ProfileSource // resolved ProfileSource implementation
}

type Style struct {
	HeaderForegroundColor   string `koanf:"headerForegroundColor" yaml:"headerForegroundColor"`
	HeaderBackgroundColor   string `koanf:"headerBackgroundColor" yaml:"headerBackgroundColor"`
	SelectedForegroundColor string `koanf:"selectedForegroundColor" yaml:"selectedForegroundColor"`
	SelectedBackgroundColor string `koanf:"selectedBackgroundColor" yaml:"selectedBackgroundColor"`
}

type CustomDebugProfile struct {
	Profiles    []Profile `koanf:"profiles" yaml:"profiles"`
	KubectlPath string    `koanf:"kubectlPath" yaml:"kubectlPath"`
	Style       Style     `koanf:"style" yaml:"style"`
}

// global Profile configuration
var Config CustomDebugProfile

func (p *Profile) IsBuiltInProfile() bool {
	return p.builtInProfile
}

func (p *Profile) SetBuiltInProfile(b bool) {
	p.builtInProfile = b
}

func (p *Profile) GetSource() ProfileSource {
	return p.source
}

func (p *Profile) SetSource(s ProfileSource) {
	p.source = s
}

// ProfileSourceConfig defines the configuration for different profile sources.
// The Type field determines which source-specific config to use.
//
//nolint:revive // ProfileSourceConfig is intentionally named this way for clarity
type ProfileSourceConfig struct {
	Type      string                 `koanf:"type" yaml:"type"`           // "file", "git", "configmap", "builtin"
	Path      string                 `koanf:"path" yaml:"path"`           // for "file" type
	Git       *GitSourceConfig       `koanf:"git" yaml:"git"`             // for "git" type
	ConfigMap *ConfigMapSourceConfig `koanf:"configMap" yaml:"configMap"` // for "configmap" type
	Name      string                 `koanf:"name" yaml:"name"`           // for "builtin" type (e.g., "netadmin")
}

// GitSourceConfig defines configuration for Git repository profile sources.
type GitSourceConfig struct {
	URL  string `koanf:"url" yaml:"url"`   // Git repository URL (e.g., "https://github.com/org/repo")
	Ref  string `koanf:"ref" yaml:"ref"`   // Branch, tag, or commit (default: "main")
	Path string `koanf:"path" yaml:"path"` // Path to profile.json within the repository
}

// ConfigMapSourceConfig defines configuration for Kubernetes ConfigMap profile sources.
type ConfigMapSourceConfig struct {
	Name string `koanf:"name" yaml:"name"` // ConfigMap name (namespace is taken from Profile.Namespace)
}
