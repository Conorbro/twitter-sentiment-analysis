package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	yaml "gopkg.in/yaml.v2"
)

// Config - struct for configuration file.
type Config struct {
	TargetHashtag                    string             `yaml:"target_hashtag"`
	TwitterCredentials               TwitterCredentials `yaml:"twitter_credentials"`
	GoogleApplicationCredentialsFile string             `yaml:"google_app_credentials_file"`

	XXX map[string]interface{} `yaml:",inline"`
}

// SafeConfig - struct containing config and mutex to allow for safe reloading.
type SafeConfig struct {
	sync.RWMutex
	C *Config
}

// ReloadConfig - allows for reloading the configuration file while the application is running.
func (sc *SafeConfig) ReloadConfig(confFile string) (err error) {
	var c = &Config{}

	yamlFile, err := ioutil.ReadFile(confFile)
	if err != nil {
		return fmt.Errorf("Error reading config file: %s", err)
	}

	if err := yaml.Unmarshal(yamlFile, c); err != nil {
		return fmt.Errorf("Error parsing config file: %s", err)
	}

	sc.Lock()
	sc.C = c
	sc.Unlock()

	return nil
}

// TwitterCredentials - struct for Twitter credentials.
type TwitterCredentials struct {
	TwitterConsumerKey       string `yaml:"twitter_consumer_key"`
	TwitterConsumerSecret    string `yaml:"twitter_consumer_secret"`
	TwitterAccessToken       string `yaml:"twitter_access_token"`
	TwitterAccessTokenSecret string `yaml:"twitter_access_token_secret"`

	XXX map[string]interface{} `yaml:",inline"`
}

func checkOverflow(m map[string]interface{}, ctx string) error {
	if len(m) > 0 {
		var keys []string
		for k := range m {
			keys = append(keys, k)
		}
		return fmt.Errorf("unknown fields in %s: %s", ctx, strings.Join(keys, ", "))
	}
	return nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (s *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Config
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	if err := checkOverflow(s.XXX, "config"); err != nil {
		return err
	}
	return nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (s *TwitterCredentials) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain TwitterCredentials
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	if err := checkOverflow(s.XXX, "config"); err != nil {
		return err
	}
	return nil
}
