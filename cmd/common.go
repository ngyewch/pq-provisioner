package cmd

import (
	"fmt"
	"github.com/ngyewch/pq-provisioner/config"
	"github.com/spf13/cobra"
)

const (
	appName = "pq-provisioner"
)

func loadConfig(cmd *cobra.Command) (*config.Main, error) {
	configFilePath, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, err
	}
	if configFilePath == "" {
		return nil, fmt.Errorf("config not provided")
	}
	cfg, err := config.LoadFromFile(configFilePath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
