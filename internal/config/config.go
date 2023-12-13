/*
Copyright 2023 The KubeArchive Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"errors"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

const (
	MetricsBindAddressKey       = "metrics.bind-address"
	HealthzBindAddressKey       = "healthz.bind-address"
	LeaderElectionKey           = "leader-election"
	CloudEventsSourceURIKey     = "cloud-events.source-uri"
	CloudEventsTargetAddressKey = "cloud-events.target-address"
	ObjectWatchesKey            = "watches"
)

// Config is a wrapper around the Viper configuration management system, with additional utility
// functions for initialization.
type Config struct {
	*viper.Viper
	initialized bool
	initOnce    sync.Once
}

func NewConfig() *Config {
	return &Config{
		Viper:       viper.New(),
		initialized: false,
	}
}

// Init initializes the configuration management system, which consists of the following:
//
// - Set default config values
// - Set up the path search for the configuration file
// - Set up environment variable aliasing
func (c *Config) Init() {
	c.initDefaults()
	c.initConfigPathSearch()
	c.initEnv()
	c.initOnce.Do(func() {
		c.initialized = true
	})
}

// IsInitialized returns true if the configuration system has been initialized.
func (c *Config) IsInitialized() bool {
	return c.initialized
}

func (c *Config) initDefaults() {
	c.SetDefault(MetricsBindAddressKey, ":8080")
	c.SetDefault(HealthzBindAddressKey, ":8081")
	c.SetDefault(LeaderElectionKey, false)
	c.SetDefault(CloudEventsSourceURIKey, "localhost")
	c.SetDefault(CloudEventsTargetAddressKey, "http://localhost:8082")
	// TODO: defaults for zap?
}

// initiConfigPathSearch initializes the search path for the configuration file. By default, it
// searches for a `dynowatch.yaml` in one of the following directories:
//
// - /etc/dynowatch
// - $HOME/.dynowatch
// - ./config/manager (below working directory)
func (c *Config) initConfigPathSearch() {
	c.SetConfigName("dynowatch")
	c.SetConfigType("yaml")
	c.AddConfigPath("/etc/dynowatch")
	c.AddConfigPath("$HOME/.dynowatch")
	c.AddConfigPath("./config/manager")
}

func (c *Config) initEnv() {
	c.SetEnvPrefix("DYNOWATCH")
	c.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	c.AutomaticEnv()
}

func (c *Config) GetWatches() ([]Watch, error) {
	watches := []Watch{}
	if !c.IsSet(ObjectWatchesKey) {
		return watches, nil
	}
	// Otherwise the data should be YAML-encoded, and we need to unmarshal...
	err := c.UnmarshalKey(ObjectWatchesKey, &watches)
	return watches, err
}

// SafeReadInConfig reads in the config file from the default search locations. It does not return
// an error if the config file is not found.
func (c *Config) SafeReadInConfig() error {
	err := c.ReadInConfig()
	if err != nil && errors.Is(err, viper.ConfigFileNotFoundError{}) {
		return nil
	}
	return err
}
