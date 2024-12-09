// SPDX-License-Identifier: MIT

package profile

import (
	corev1 "k8s.io/api/core/v1"
)

type Profile struct {
	ProfileName     string            `koanf:"name" yaml:"name" validate:"required"`
	Profile         string            `koanf:"profile" yaml:"profile" validate:"required"`
	Image           string            `koanf:"image" yaml:"image" validate:"required"`
	Namespace       string            `koanf:"namespace" yaml:"namespace" validate:"required"`
	ImagePullPolicy corev1.PullPolicy `koanf:"imagePullPolicy" yaml:"imagePullPolicy" validate:"required"`
	TargetContainer string            `koanf:"targetContainer" yaml:"targetContainer" validate:"required"`
	MatchLabels     map[string]string `koanf:"matchLabels" yaml:"matchLabels" validate:"required"`

	// only used internally
	builtInProfile bool
}

type CustomDebugProfile struct {
	Profiles    []Profile `koanf:"profiles" yaml:"profiles"`
	KubectlPath string    `koanf:"kubectlPath" yaml:"kubectlPath"`
}

// global Profile configuration
var Config CustomDebugProfile

func (p *Profile) IsBuiltInProfile() bool {
	return p.builtInProfile
}

func (p *Profile) SetBuiltInProfile(b bool) {
	p.builtInProfile = b
}
