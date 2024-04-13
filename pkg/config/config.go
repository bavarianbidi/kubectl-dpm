// SPDX-License-Identifier: MIT

package config

import (
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/pkg/errors"

	"github.com/bavarianbidi/kubectl-dpm/pkg/profile"
)

// ConfigurationFile is the path to the configuration file
// is getting set by the root command as flag with default to ~/.kube/debug-profiles.yaml
var ConfigurationFile string

func GenerateConfig() error {
	k := koanf.New(".")

	// load config from file if given or from default paths
	if err := k.Load(file.Provider(ConfigurationFile), yaml.Parser()); err != nil {
		return errors.Wrap(err, "failed to load config file")
	}

	// unmarshal all koanf config keys into the global Config struct
	if err := k.Unmarshal("", &profile.Config); err != nil {
		return errors.Wrap(err, "failed to unmarshal config")
	}

	return nil
}
