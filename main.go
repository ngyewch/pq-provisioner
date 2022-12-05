package main

import (
	"github.com/ngyewch/pq-provisioner/cmd"
	"github.com/ngyewch/pq-provisioner/common"
	goVersion "go.hein.dev/go-version"
)

var (
	version string
	commit  string
	date    string
)

func main() {
	versionInfo := goVersion.New(version, commit, date)
	common.VersionInfo = versionInfo

	cmd.Execute()
}
