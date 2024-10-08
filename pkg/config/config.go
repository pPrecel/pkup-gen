package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// path to the report template
	Template string `yaml:"template"`
	// repos based on which report will be generated ( with name in format <ORG>/<REPO> )
	// can override orgs config for a specific repo
	Repos []Remote `yaml:"repos,omitempty"`
	// orgs based on which report will be generated ( with name in format <ORG> )
	Orgs []Remote `yaml:"orgs,omitempty"`
	// info about output reports
	Reports []Report `yaml:"reports,omitempty"`
	// info about email server used to send emails
	Send Send `yaml:"send,omitempty"`
}

type Send struct {
	// email server address
	// e.g.: "smtp-mail.outlook.com"
	ServerAddress string `yaml:"serverAddress"`
	// email server port
	// e.g.: 587
	ServerPort int `yaml:"serverPort"`
	// email server username
	// e.g.: filip.strozik@outlook.com
	Username string `yaml:"username"`
	// email server password
	// how to create app password for gmail:
	// https://support.google.com/accounts/answer/185833?hl=en
	// e.g.: testpassword
	Password string `yaml:"password"`
	// how many emails should be send on single dial ( default: 1 )
	// gmail smtp server limitations:
	// https://support.google.com/a/answer/166852?hl=en
	// e.g.: 30
	PerDial int `yaml:"perDial,omitempty"`
	// delay between dials ( default: 0s )
	// e.g.: 60s
	DialDelay time.Duration `yaml:"dialDelay,omitempty"`
	// message subject
	// e.g.: "PKUP report"
	Subject string `yaml:"subject"`
	// message body html template path
	// e.g.: template.html
	HTMLBodyPath string `yaml:"htmlBodyPath,omitempty"`
	// message author
	// e.g.: filip.strozik@outlook.com
	From string `yaml:"from"`
}

type Remote struct {
	// name of the remot ( in format <ORG> for orgs or <ORG>/<REPO> for repos )
	// e.g.: "kyma-project" or "kyma-project/serverless"
	Name string `yaml:"name"`
	// token used to communicate with the GitHub API
	Token string `yaml:"token,omitempty"`
	// enterprise GitHub API address ( default: use opensource GitHub API address )
	EnterpriseUrl string `yaml:"enterpriseUrl,omitempty"`
	// specific branches used to fetch commits from ( default: use repo HEAD branch )
	Branches []string `yaml:"branches,omitempty"`
	// fetch commits from all branches instead of HEAD branch ( default: false )
	AllBranches bool `yaml:"allBranches"`
	// fetch only unique commits ( default: false )
	// mostly useful in case the Branches or AllBranches variable is set
	UniqueOnly bool `yaml:"uniqueOnly"`
}

type Report struct {
	// set of GitHub usernames that report will be based on
	Signatures []Signature `yaml:"signatures,omitempty"`
	// report owner's email address used to send mail
	Email string `yaml:"email,omitempty"`
	// output dir where report will be generated
	OutputDir string `yaml:"outputDir,omitempty"`
	// extra fields that will be replaces in the template report
	// e.g.: pkupGenEmployeesName: "Filip Strózik"
	ExtraFields map[string]string `yaml:"extraFields,omitempty"`
}

type Signature struct {
	// GitHub username
	Username string `yaml:"username"`
	// enterprise GitHub API address of correlated Remote
	// if empty uses username for all remotes with no EnterpriseUrl set
	// if not empty uses username for all remotes with the same EnterpriseUrl
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
