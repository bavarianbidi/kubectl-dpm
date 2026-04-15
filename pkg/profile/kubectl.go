// SPDX-License-Identifier: MIT

package profile

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	kubectlversion "k8s.io/kubectl/pkg/cmd/version"
)

func CheckKubectlVersion() error {
	// nolint:gosec
	kubeclVersionCmd := exec.Command(
		os.ExpandEnv(Config.KubectlPath),
		"version",
		"--client", "true",
		"--output", "json",
	)

	kubectlOutput, err := kubeclVersionCmd.Output()
	if err != nil {
		return fmt.Errorf("get kubectl version output: %w", err)
	}

	kubectlVersion := &kubectlversion.Version{}

	err = json.Unmarshal(kubectlOutput, kubectlVersion)
	if err != nil {
		return fmt.Errorf("parse kubectl version output: %w", err)
	}

	kubectlMinorVersion, err := strconv.Atoi(kubectlVersion.ClientVersion.Minor)
	if err != nil {
		return fmt.Errorf("parse kubectl minor version %q: %w", kubectlVersion.ClientVersion.Minor, err)
	}

	if kubectlMinorVersion < 30 {
		return fmt.Errorf("your kubectl doesn't support custom debug profiles, please upgrade to kubectl 1.30 or newer or use v0.0.4 of kubectl-dpm")
	}

	return nil
}
