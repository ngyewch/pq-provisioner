package cmd

import (
	"fmt"
	"github.com/ngyewch/go-clibase"
	"github.com/ngyewch/pq-provisioner/common"
	"github.com/spf13/cobra"
	goVersion "go.hein.dev/go-version"
	"os"
)

var (
	rootCmd = &cobra.Command{
		Use:   fmt.Sprintf("%s [flags]", appName),
		Short: "spd CLI",
		RunE:  help,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func help(cmd *cobra.Command, args []string) error {
	err := cmd.Help()
	if err != nil {
		return err
	}
	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("config", "", "Config file")

	clibase.AddVersionCmd(rootCmd, func() *goVersion.Info {
		return common.VersionInfo
	})
}

func initConfig() {
}
