// SPDX-License-Identifier: MIT

package profile

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"slices"
	"strings"

	corev1 "k8s.io/api/core/v1"
	kubectldebug "k8s.io/kubectl/pkg/cmd/debug"
)

func ValidateDebugProfileFile() error {
	if err := ValidateKubectlPath(); err != nil {
		return err
	}

	if err := ValidateAllProfiles(); err != nil {
		return err
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
			return err
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
			return fmt.Errorf("profile %s is missing a custom profile name", p.Profile)
		case p.Profile == "":
			Config.Profiles = slices.Delete(Config.Profiles, idx, idx)
			return fmt.Errorf("profile name %s is either missing a profile file or the name of a built-in profile", p.ProfileName)
		}

		if err := ValidateProfile(p.ProfileName); err != nil {
			return err
		}
	}
	return nil
}

// ValidateProfile validates a single profile
func ValidateProfile(profileName string) error {
	idx := GetProfileIdx(profileName)

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
func InteractiveProfiles() []Profile {
	// sort profiles by name
	SortProfiles()

	ValidateAllProfiles()

	var interactiveProfiles []Profile

	for _, p := range Config.Profiles {
		if !namespaceIsMissing(p.ProfileName) &&
			!labelSelectorIsMissing(p.ProfileName) &&
			!imageIsMissing(p.ProfileName) {
			interactiveProfiles = append(interactiveProfiles, p)
		}
	}

	return interactiveProfiles
}

func namespaceIsMissing(profileName string) bool {
	return Config.Profiles[GetProfileIdx(profileName)].Namespace == ""
}

func labelSelectorIsMissing(profileName string) bool {
	return Config.Profiles[GetProfileIdx(profileName)].MatchLabels == nil
}

func imageIsMissing(profileName string) bool {
	return Config.Profiles[GetProfileIdx(profileName)].Image == ""
}

func validatePodSpec(podSpec string) error {
	podSpecByte, err := os.ReadFile(os.ExpandEnv(podSpec))
	if err != nil {
		return err
	}

	pod := corev1.PodSpec{}

	if err := json.Unmarshal(podSpecByte, &pod); err != nil {
		return err
	}

	return nil
}

// CompleteProfile completes a profile with default values
func CompleteProfile(profileName string) error {
	// get the index of the profile where the profile name matches
	idx := GetProfileIdx(profileName)

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
