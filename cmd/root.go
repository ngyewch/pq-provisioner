package cmd

import (
	"fmt"
	versionInfoCobra "github.com/ngyewch/go-versioninfo/cobra"
	"github.com/spf13/cobra"
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

	versionInfoCobra.AddVersionCmd(rootCmd, nil)
}

func initConfig() {
}
