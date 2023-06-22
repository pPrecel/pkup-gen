package cmd

import (
	"github.com/sirupsen/logrus"
)

type Options struct {
	Version string
	Log     *logrus.Logger
}

type genActionOpts struct {
	*Options
	perdiod       int
	dir           string
	repos         map[string][]string
	username      string
	token         string
	enterpriseURL string
}
