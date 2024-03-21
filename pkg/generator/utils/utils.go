package utils

import "strings"

func SplitRemoteName(remote string) (string, string) {
	repoOrg := strings.Split(remote, "/")
	return repoOrg[0], repoOrg[1]
}
