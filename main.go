package main

import (
	"context"
	"log"
	"os"

	"github.com/ngyewch/pq-provisioner/config"
	"github.com/ngyewch/pq-provisioner/provisioner"
	"github.com/urfave/cli/v3"
)

var (
	version string

	flagConfig = &cli.StringFlag{
		Name:     "config",
		Usage:    "config file",
		Required: true,
	}

	app = &cli.Command{
		Name:    "pq-provisioner",
		Usage:   "PostgreSQL provisioner",
		Version: version,
		Action:  nil,
		Commands: []*cli.Command{
			{
				Name:   "provision",
				Usage:  "provision",
				Action: doProvision,
				Flags: []cli.Flag{
					flagConfig,
				},
			},
		},
	}
)

func main() {
	err := app.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func doProvision(ctx context.Context, cmd *cli.Command) error {
	configFilePath := cmd.String(flagConfig.Name)

	cfg, err := config.LoadFromFile(configFilePath)
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
