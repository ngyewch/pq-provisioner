package cmd

import (
	"fmt"
	slog "github.com/go-eden/slf4go"
	"github.com/ngyewch/pq-provisioner/config"
	"github.com/spf13/cobra"
	"os"
)

const (
	appName = "pq-provisioner"
)

var (
	log = slog.GetLogger()
)

func loadConfig(cmd *cobra.Command) (*config.Main, error) {
	configFilePath, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, err
	}
	if configFilePath == "" {
		return nil, fmt.Errorf("config not provided")
	}
	_, err = os.Stat(configFilePath)
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load(configFilePath)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
