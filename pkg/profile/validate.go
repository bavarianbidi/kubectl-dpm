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
			log.Printf("duplicate profile name %s found - keep the one with profile file %s\n", a.ProfileName, a.CustomProfileFile)
			return true
		}
		return false
	})

	// update Config.Profiles with compacted profiles
	Config.Profiles = compactProfiles

	for _, p := range Config.Profiles {
		switch {
		case p.ProfileName == "":
			return fmt.Errorf("profile file %s is missing a profile name", p.CustomProfileFile)
		case p.CustomProfileFile == "":
			return fmt.Errorf("profile name %s is missing a profile file", p.ProfileName)
		}

		if err := ValidateProfile(p.ProfileName); err != nil {
			return err
		}
	}
	return nil
}

// ValidateProfile validates a single profile
func ValidateProfile(profileName string) error {
	idx := slices.IndexFunc(Config.Profiles,
		func(c Profile) bool { return c.ProfileName == profileName },
	)

	if err := validatePodSpec(Config.Profiles[idx].CustomProfileFile); err != nil {
		return err
	}

	return nil
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

func ValidateAndCompleteProfile(profileName string) error {
	if err := CompleteProfile(profileName); err != nil {
		return err
	}

	if err := ValidateProfile(profileName); err != nil {
		return err
	}
	return nil
}

// CompleteProfile completes a profile with default values
func CompleteProfile(profileName string) error {
	// get the index of the profile where the profile name matches
	idx := slices.IndexFunc(Config.Profiles,
		func(c Profile) bool { return c.ProfileName == profileName },
	)

	if Config.Profiles[idx].ImagePullPolicy == "" {
		Config.Profiles[idx].ImagePullPolicy = corev1.PullIfNotPresent
	}

	return nil
}
