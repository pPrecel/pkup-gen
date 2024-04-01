package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Template string   `yaml:"template"`
	Repos    []Remote `yaml:"repos,omitempty"`
	Orgs     []Remote `yaml:"orgs,omitempty"`
	Reports  []Report `yaml:"reports,omitempty"`
	Send     Send     `yaml:"send,omitempty"`
}

type Send struct {
	ServerAddress string         `yaml:"serverAddress"`
	ServerPort    int            `yaml:"serverPort"`
	Username      string         `yaml:"username"`
	Password      string         `yaml:"password"`
	Delay         *time.Duration `yaml:"delay,omitempty"`
	Subject       string         `yaml:"subject"`
	HTMLBodyPath  string         `yaml:"htmlBodyPath,omitempty"`
	From          string         `yaml:"from"`
}

type Remote struct {
	Name          string   `yaml:"name"`
	Token         string   `yaml:"token,omitempty"`
	EnterpriseUrl string   `yaml:"enterpriseUrl,omitempty"`
	Branches      []string `yaml:"branches,omitempty"`
	AllBranches   bool     `yaml:"allBranches"`
	UniqueOnly    bool     `yaml:"uniqueOnly"`
}

type Report struct {
	Signatures  []Signature       `yaml:"signatures,omitempty"`
	Email       string            `yaml:"email,omitempty"`
	OutputDir   string            `yaml:"outputDir,omitempty"`
	ExtraFields map[string]string `yaml:"extraFields,omitempty"`
}

type Signature struct {
	Username      string `yaml:"username"`
	EnterpriseUrl string `yaml:"enterpriseUrl,omitempty"`
}

func Read(path string) (*Config, error) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}

	return config, yaml.Unmarshal(yamlFile, config)
}
