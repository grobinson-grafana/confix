package main

import (
	"fmt"
	"github.com/go-test/deep"
	"github.com/grafana/mimir/pkg/alertmanager/alertspb"
	"github.com/prometheus/alertmanager/config"
	"gopkg.in/yaml.v2"
)

// cfg is a special version of an Alertmanager configuration that uses yaml.MapSlice
// to avoid issues when deserializing and re-serializing Alertmanager configurations,
// such as secrets being replaced with <secret>. We use this struct to double quote
// all matchers in a configuration file.
type cfg struct {
	Global            yaml.MapSlice        `yaml:"global,omitempty" json:"global,omitempty"`
	Route             *config.Route        `yaml:"route,omitempty" json:"route,omitempty"`
	InhibitRules      []config.InhibitRule `yaml:"inhibit_rules,omitempty" json:"inhibit_rules,omitempty"`
	Receivers         []yaml.MapSlice      `yaml:"receivers,omitempty" json:"receivers,omitempty"`
	Templates         []string             `yaml:"templates" json:"templates"`
	MuteTimeIntervals []yaml.MapSlice      `yaml:"mute_time_intervals,omitempty" json:"mute_time_intervals,omitempty"`
	TimeIntervals     []yaml.MapSlice      `yaml:"time_intervals,omitempty" json:"time_intervals,omitempty"`
}

// fix double quotes all matchers in a configuration file. It uses the fact that MarshalYAML for
// labels.Matchers double quotes matchers, even if the original is unquoted. It does not check
// if the fixed configuration is equivalent to the original configuration. Use the isEqual
// function for this.
func fix(desc alertspb.AlertConfigDesc) (*alertspb.AlertConfigDesc, error) {
	tmp := cfg{}
	if err := yaml.Unmarshal([]byte(desc.RawConfig), &tmp); err != nil {
		return nil, fmt.Errorf("failed to load config from YAML: %w", err)
	}
	b, err := yaml.Marshal(tmp)
	if err != nil {
		return nil, fmt.Errorf("failed to save config as YAML: %w", err)
	}
	res := desc
	res.RawConfig = string(b)
	return &res, nil
}

// isEqual returns true if both configurations are equivalent. An example of two configurations
// that are different, but equivalent, are those with fields in different orders, quoted v.s.
// unquoted YAML strings, etc. It returns false and a list of differences if the configurations
// are not equal, and an error if the function could not compare the two configurations.
func isEqual(desc1, desc2 alertspb.AlertConfigDesc) (bool, []string, error) {
	tmp1, err := config.Load(desc1.RawConfig)
	if err != nil {
		return false, nil, fmt.Errorf("failed to load config in desc1 from YAML: %w", err)
	}
	tmp2, err := config.Load(desc2.RawConfig)
	if err != nil {
		return false, nil, fmt.Errorf("failed to load config in desc2 from YAML: %w", err)
	}
	if diffs := deep.Equal(tmp1, tmp2); diffs != nil {
		return false, diffs, nil
	}
	return true, nil, nil
}
