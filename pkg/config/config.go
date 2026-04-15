// SPDX-License-Identifier: MIT

package config

import (
	"fmt"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"

	"github.com/bavarianbidi/kubectl-dpm/pkg/profile"
)

// ConfigurationFile is the path to the configuration file
// is getting set by the root command as flag with default to ~/.kube/debug-profiles.yaml
var ConfigurationFile string

func GenerateConfig() error {
	k := koanf.New(".")

	// load config from file if given or from default paths
	if err := k.Load(file.Provider(ConfigurationFile), yaml.Parser()); err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	// unmarshal all koanf config keys into the global Config struct
	if err := k.Unmarshal("", &profile.Config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
