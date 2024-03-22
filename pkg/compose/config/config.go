package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Template string   `yaml:"template"`
	Repos    []Remote `yaml:"repos,omitempty"`
	Orgs     []Remote `yaml:"orgs,omitempty"`
	Users    []User   `yaml:"users,omitempty"`
}

type Remote struct {
	Name          string   `yaml:"name"`
	Token         string   `yaml:"token,omitempty"`
	EnterpriseUrl string   `yaml:"enterpriseUrl,omitempty"`
	Branches      []string `yaml:"branches,omitempty"`
	AllBranches   bool     `yaml:"allBranches"`
	UniqueOnly    bool     `yaml:"uniqueOnly"`
}

type User struct {
	Usernames    []Username        `yaml:"usernames,omitempty"`
	OutputDir    string            `yaml:"outputDir,omitempty"`
	ReportFields map[string]string `yaml:"reportFields,omitempty"`
}

type Username struct {
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
