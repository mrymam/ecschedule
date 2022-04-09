package ecschedule

import (
	"io"
	"io/ioutil"
	"text/template"

	"github.com/goccy/go-yaml"
	gc "github.com/kayac/go-config"
)

const defaultRole = "ecsEventsRole"

// BaseConfig baseconfig
type BaseConfig struct {
	Region    string `yaml:"region"`
	Cluster   string `yaml:"cluster"`
	AccountID string `yaml:"-"`
}

// Config config
type Config struct {
	Role        string `yaml:"role,omitempty"`
	*BaseConfig `yaml:",inline"`
	Rules       []*Rule   `yaml:"rules"`
	Plugins     []*Plugin `yaml:"plugins,omitempty"`

	templateFuncs []template.FuncMap
}

// GetRuleByName gets rule by name
func (c *Config) GetRuleByName(name string) *Rule {
	for _, r := range c.Rules {
		if r.Name == name {
			return r
		}
	}
	return nil
}

func (c *Config) setupPlugins() error {
	for _, p := range c.Plugins {
		if err := p.Setup(c); err != nil {
			return err
		}
	}
	return nil
}

// LoadConfig loads config
func LoadConfig(r io.Reader, accountID string) (*Config, error) {
	c := Config{}
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	bs, err = envReplacer(bs)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(bs, &c); err != nil {
		return nil, err
	}
	c.AccountID = accountID
	if err := c.setupPlugins(); err != nil {
		return nil, err
	}
	loader := gc.New()
	for _, f := range c.templateFuncs {
		loader.Funcs(f)
	}
	for _, r := range c.Rules {
		r.mergeBaseConfig(c.BaseConfig, c.Role)
	}
	return &c, nil
}
