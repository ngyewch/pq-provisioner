package main

import (
	"fmt"
	"github.com/ngyewch/pq-provisioner/config"
	"github.com/ngyewch/pq-provisioner/provisioner"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	version         string
	commit          string
	commitTimestamp string

	flagConfig = &cli.PathFlag{
		Name:     "config",
		Usage:    "config file",
		Required: true,
	}

	app = &cli.App{
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
	cli.VersionPrinter = func(cCtx *cli.Context) {
		var parts []string
		if version != "" {
			parts = append(parts, fmt.Sprintf("version=%s", version))
		}
		if commit != "" {
			parts = append(parts, fmt.Sprintf("commit=%s", commit))
		}
		if commitTimestamp != "" {
			formattedCommitTimestamp := func(commitTimestamp string) string {
				epochSeconds, err := strconv.ParseInt(commitTimestamp, 10, 64)
				if err != nil {
					return ""
				}
				t := time.Unix(epochSeconds, 0)
				return t.Format(time.RFC3339)
			}(commitTimestamp)
			if formattedCommitTimestamp != "" {
				parts = append(parts, fmt.Sprintf("commitTimestamp=%s", formattedCommitTimestamp))
			}
		}
		fmt.Println(strings.Join(parts, " "))
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func doProvision(cCtx *cli.Context) error {
	configFilePath := flagConfig.Get(cCtx)

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
