package cmd

import (
	"github.com/ngyewch/pq-provisioner/provisioner"
	"github.com/spf13/cobra"
)

var (
	provisionCmd = &cobra.Command{
		Use:   "provision",
		Short: "Provision",
		Args:  cobra.ExactArgs(0),
		RunE:  provision,
	}
)

func provision(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	configProvisioner, err := provisioner.NewConfigProvisioner(cfg, nil)
	if err != nil {
		return err
	}

	err = configProvisioner.Provision()
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(provisionCmd)

	provisionCmd.Flags().String("config", "", "Config file")
}
