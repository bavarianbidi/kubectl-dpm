// SPDX-License-Identifier: MIT

package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"slices"
	"strings"

	corev1 "k8s.io/api/core/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	kubectldebug "k8s.io/kubectl/pkg/cmd/debug"
)

func ValidateDebugProfileFile() error {
	if err := ValidateKubectlPath(); err != nil {
		return fmt.Errorf("validate kubectl path: %w", err)
	}

	if err := ValidateAllProfiles(); err != nil {
		return fmt.Errorf("validate all profiles: %w", err)
	}

	return nil
}

func ValidateKubectlPath() error {
	// if plugin is called from kubectl, we do not
	// check the kubectl path
	r := regexp.MustCompile(".*kubectl$")

	if r.MatchString(os.Getenv("_")) {
		Config.KubectlPath = os.Getenv("_")
	} else {
		// check if configured kubectl path is valid
		info, err := os.Stat(os.ExpandEnv(Config.KubectlPath))
		if os.IsNotExist(err) {
			return fmt.Errorf("kubectl %s does not exist", Config.KubectlPath)
		}
		if info.IsDir() {
			return fmt.Errorf("kubectl path %q is a directory, not an executable", Config.KubectlPath)
		}

		mode := info.Mode()
		if (mode & 0o111) == 0 {
			return os.ErrPermission
		}
	}
	return nil
}

func ValidateAllProfiles() error {
	// sort profiles by name
	SortProfiles()

	compactProfiles := slices.CompactFunc(Config.Profiles, func(a, b Profile) bool {
		if strings.EqualFold(a.ProfileName, b.ProfileName) {
			log.Printf("duplicate profile name %s found - keep the one with profile file %s\n", a.ProfileName, a.Profile)
			return true
		}
		return false
	})

	// update Config.Profiles with compacted profiles
	Config.Profiles = compactProfiles

	for idx, p := range Config.Profiles {
		switch {
		case p.ProfileName == "":
			Config.Profiles = slices.Delete(Config.Profiles, idx, idx)
			return fmt.Errorf("profile at index %d is missing a custom profile name", idx)
		case p.ProfileSource.Type == "" && p.Profile == "":
			Config.Profiles = slices.Delete(Config.Profiles, idx, idx)
			return fmt.Errorf("profile %q is missing both profileSource and profile (legacy) configuration", p.ProfileName)
		}

		if err := ValidateProfile(p.ProfileName); err != nil {
			return fmt.Errorf("validate profile %q: %w", p.ProfileName, err)
		}
	}
	return nil
}

// ValidateProfile validates a single profile and instantiates its ProfileSource
func ValidateProfile(profileName string) error {
	idx, err := GetProfileIdx(profileName)
	if err != nil {
		return err
	}
	profile := &Config.Profiles[idx]

	// Check if using new ProfileSource config or legacy Profile field
	if profile.ProfileSource.Type != "" {
		// New ProfileSource configuration
		return validateAndInstantiateProfileSource(profile, nil)
	}

	// Legacy Profile field - handle for backward compatibility
	if profile.Profile == "" {
		return fmt.Errorf("profile %q is missing both profileSource and profile fields", profileName)
	}

	return validateLegacyProfile(idx)
}

// validateAndInstantiateProfileSource creates the appropriate ProfileSource implementation
// based on the ProfileSourceConfig. The k8sClient parameter is optional and only needed
// for ConfigMap sources.
func validateAndInstantiateProfileSource(p *Profile, k8sClient corev1client.CoreV1Interface) error {
	var source ProfileSource
	var err error

	switch p.ProfileSource.Type {
	case SourceTypeFile:
		if p.ProfileSource.Path == "" {
			return fmt.Errorf("file profile source requires 'path' field")
		}
		source = NewFileProfileSource(p.ProfileSource.Path)

	case SourceTypeBuiltIn:
		if p.ProfileSource.Name == "" {
			return fmt.Errorf("builtin profile source requires 'name' field")
		}
		source, err = NewBuiltInProfileSource(p.ProfileSource.Name)
		if err != nil {
			return fmt.Errorf("create builtin profile source: %w", err)
		}
		p.SetBuiltInProfile(true)

	case SourceTypeGit:
		if p.ProfileSource.Git == nil {
			return fmt.Errorf("git profile source requires 'git' configuration")
		}
		if p.ProfileSource.Git.URL == "" {
			return fmt.Errorf("git profile source requires 'git.url' field")
		}
		if p.ProfileSource.Git.Path == "" {
			return fmt.Errorf("git profile source requires 'git.path' field")
		}
		source = NewGitProfileSource(
			p.ProfileSource.Git.URL,
			p.ProfileSource.Git.Ref,
			p.ProfileSource.Git.Path,
		)

	case SourceTypeConfigMap:
		if p.ProfileSource.ConfigMap == nil {
			return fmt.Errorf("configmap profile source requires 'configMap' configuration")
		}
		if p.ProfileSource.ConfigMap.Name == "" {
			return fmt.Errorf("configmap profile source requires 'configMap.name' field")
		}
		if p.Namespace == "" {
			return fmt.Errorf("configmap profile source requires 'namespace' field to locate the ConfigMap")
		}
		// For now, we'll create the source without a client during validation
		// The actual client will be injected later when needed
		if k8sClient != nil {
			source = NewConfigMapProfileSource(k8sClient, p.Namespace, p.ProfileSource.ConfigMap.Name)
		} else {
			// Validation-only mode - we can't actually fetch the ConfigMap yet
			// Just verify the configuration is complete
			log.Printf("configmap profile source %q will be validated at runtime\n", p.ProfileName)
			return nil
		}

	default:
		return fmt.Errorf("unknown profile source type: %q (valid types: file, git, configmap, builtin)", p.ProfileSource.Type)
	}

	// Store the source implementation
	p.SetSource(source)

	// Validate by fetching the spec (except for ConfigMap during initial validation)
	if p.ProfileSource.Type != SourceTypeConfigMap {
		_, err = source.GetSpec(context.Background())
		if err != nil {
			log.Printf("profile %s validation warning: %s\n", p.ProfileName, err.Error())
		}
	}

	return nil
}

// validateLegacyProfile validates profiles using the legacy 'profile' field
func validateLegacyProfile(idx int) error {
	switch Config.Profiles[idx].Profile {
	case kubectldebug.ProfileLegacy:
		Config.Profiles[idx].SetBuiltInProfile(true)
		return nil
	case kubectldebug.ProfileGeneral:
		Config.Profiles[idx].SetBuiltInProfile(true)
		return nil
	case kubectldebug.ProfileBaseline:
		Config.Profiles[idx].SetBuiltInProfile(true)
		return nil
	case kubectldebug.ProfileRestricted:
		Config.Profiles[idx].SetBuiltInProfile(true)
		return nil
	case kubectldebug.ProfileNetadmin:
		Config.Profiles[idx].SetBuiltInProfile(true)
		return nil
	case kubectldebug.ProfileSysadmin:
		Config.Profiles[idx].SetBuiltInProfile(true)
		return nil
	default:
		if err := validatePodSpec(Config.Profiles[idx].Profile); err != nil {
			log.Printf("profile %s is invalid: %s\n", Config.Profiles[idx].Profile, err.Error())
		}
	}

	return nil
}

// InteractiveProfiles returns all profiles that can be used from
// the interactive mode, dpm run (without args)
// these profiles do have all required fields set (namespace, labelSelector, image)
func InteractiveProfiles() ([]Profile, error) {
	// sort profiles by name
	SortProfiles()

	err := ValidateAllProfiles()
	if err != nil {
		return nil, fmt.Errorf("validate all profiles: %w", err)
	}

	var interactiveProfiles []Profile

	for _, p := range Config.Profiles {
		if !namespaceIsMissing(p.ProfileName) &&
			!labelSelectorIsMissing(p.ProfileName) &&
			!imageIsMissing(p.ProfileName) {
			interactiveProfiles = append(interactiveProfiles, p)
		}
	}

	return interactiveProfiles, nil
}

func namespaceIsMissing(profileName string) bool {
	idx, err := GetProfileIdx(profileName)
	if err != nil {
		return true
	}
	return Config.Profiles[idx].Namespace == ""
}

func labelSelectorIsMissing(profileName string) bool {
	idx, err := GetProfileIdx(profileName)
	if err != nil {
		return true
	}
	return Config.Profiles[idx].MatchLabels == nil
}

func imageIsMissing(profileName string) bool {
	idx, err := GetProfileIdx(profileName)
	if err != nil {
		return true
	}
	return Config.Profiles[idx].Image == ""
}

func validatePodSpec(podSpec string) error {
	podSpecByte, err := os.ReadFile(os.ExpandEnv(podSpec))
	if err != nil {
		return fmt.Errorf("read profile file %q: %w", podSpec, err)
	}

	pod := corev1.PodSpec{}

	if err := json.Unmarshal(podSpecByte, &pod); err != nil {
		return fmt.Errorf("parse profile JSON from %q: %w", podSpec, err)
	}

	return nil
}

// CompleteProfile completes a profile with default values
func CompleteProfile(profileName string) error {
	// get the index of the profile where the profile name matches
	idx, err := GetProfileIdx(profileName)
	if err != nil {
		return err
	}

	if Config.Profiles[idx].ImagePullPolicy == "" {
		Config.Profiles[idx].ImagePullPolicy = corev1.PullIfNotPresent
	}

	return nil
}

// CompleteStyle completes the style with default values
func CompleteStyle() {
	if Config.Style.HeaderForegroundColor == "" {
		Config.Style.HeaderForegroundColor = "#ffffaf"
	}
	if Config.Style.HeaderBackgroundColor == "" {
		Config.Style.HeaderBackgroundColor = "#5f00ff"
	}
	if Config.Style.SelectedForegroundColor == "" {
		Config.Style.SelectedForegroundColor = "#ffffaf"
	}
	if Config.Style.SelectedBackgroundColor == "" {
		Config.Style.SelectedBackgroundColor = "#5f00ff"
	}
}

// InitializeConfigMapSource creates and sets a ConfigMapProfileSource with the given Kubernetes client.
// This function must be called at runtime before using a ConfigMap profile source.
func InitializeConfigMapSource(p *Profile, client corev1client.CoreV1Interface) error {
	if p.ProfileSource.Type != SourceTypeConfigMap {
		return fmt.Errorf("profile is not a configmap source")
	}

	if p.ProfileSource.ConfigMap == nil || p.ProfileSource.ConfigMap.Name == "" {
		return fmt.Errorf("configmap source configuration is missing")
	}

	source := NewConfigMapProfileSource(client, p.Namespace, p.ProfileSource.ConfigMap.Name)
	p.SetSource(source)

	// Validate by fetching the spec
	_, err := source.GetSpec(context.Background())
	if err != nil {
		return fmt.Errorf("validate configmap source: %w", err)
	}

	return nil
}
