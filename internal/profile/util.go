// SPDX-License-Identifier: MIT

package profile

import (
	"cmp"
	"fmt"
	"slices"
	"strings"
)

// SortProfiles sorts the global profiles list alphabetically by profile name.
func SortProfiles() {
	// sort profiles by name
	slices.SortFunc(Config.Profiles, func(a, b Profile) int {
		return cmp.Compare(strings.ToLower(a.ProfileName), strings.ToLower(b.ProfileName))
	})
}

// GetProfileIdx returns the index of a profile matching the given profile name.
func GetProfileIdx(profileName string) (int, error) {
	// get the index of the profile where the profile name matches
	idx := slices.IndexFunc(Config.Profiles, func(c Profile) bool { return c.ProfileName == profileName })
	if idx == -1 {
		return -1, fmt.Errorf("profile %q not found", profileName)
	}
	return idx, nil
}
