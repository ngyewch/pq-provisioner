package cmd

import (
	"database/sql"
	"fmt"
	"github.com/ngyewch/pq-provisioner/config"
	"github.com/ngyewch/pq-provisioner/provisioner"
	sshTunnel "github.com/ngyewch/pq-provisioner/ssh-tunnel"
	"github.com/spf13/cobra"
	"log/slog"
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
	log := slog.Default()

	cfg, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	dbHost := cfg.Host
	dbPort := cfg.Port
	if dbPort == 0 {
		dbPort = 5432
	}

	if cfg.Ssh != nil {
		client, err := sshTunnel.NewSshClient(cfg.Ssh.Host, cfg.Ssh.Port, cfg.Ssh.User, cfg.Ssh.IdentityFile)
		if err != nil {
			return err
		}
		defer client.Close()

		lpf, err := sshTunnel.NewLocalPortForwarder(client, "localhost:0", fmt.Sprintf("%s:%d", dbHost, dbPort))
		if err != nil {
			return err
		}
		defer lpf.Close()

		lpf.Start()

		dbHost = "localhost"
		dbPort = lpf.LocalAddr().Port
	}

	connStr := buildConnectionString(cfg, cfg.Database, cfg.User, dbHost, dbPort)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	prov, err := provisioner.NewProvisioner(db)
	if err != nil {
		return err
	}

	for _, database := range cfg.Databases {
		if !prov.HasDatabase(database.Name) {
			log.Info(fmt.Sprintf("Creating database %s...", database.Name))
			err := prov.CreateDatabase(database.Name)
			if err != nil {
				return err
			}
		}
		if !prov.HasUser(database.Owner) {
			log.Info(fmt.Sprintf("Creating user %s...", database.Owner))
			user := cfg.GetUser(database.Owner)
			if user == nil {
				return fmt.Errorf("user not defined")
			}
			if user.Password == "" {
				return fmt.Errorf("user password not specified")
			}
			err := prov.CreateUser(user.Name, user.Password)
			if err != nil {
				return err
			}
		}
		if !prov.HasUser(database.User) {
			log.Info(fmt.Sprintf("Creating user %s...", database.User))
			user := cfg.GetUser(database.User)
			if user == nil {
				return fmt.Errorf("user not defined")
			}
			if user.Password == "" {
				return fmt.Errorf("user password not specified")
			}
			err := prov.CreateUser(user.Name, user.Password)
			if err != nil {
				return err
			}
		}
		log.Info(fmt.Sprintf("Setting database %s owner to %s...", database.Name, database.Owner))
		err = prov.SetDatabaseOwner(database.Name, database.Owner)
		if err != nil {
			return err
		}

		connStr := buildConnectionString(cfg, database.Name, database.Owner, dbHost, dbPort)
		db2, err := sql.Open("postgres", connStr)
		if err != nil {
			return err
		}
		defer db2.Close()

		log.Info(fmt.Sprintf("Setting database %s user to %s...", database.Name, database.User))
		err = provisioner.SetDatabaseUser(db2, database.User)
		if err != nil {
			return err
		}
	}

	return nil
}

func buildConnectionString(cfg *config.Main, dbname string, user string, host string, port int) string {
	userEntry := cfg.GetUser(user)
	password := ""
	if userEntry != nil {
		password = userEntry.Password
	}
	if (host == "") && (password == "") {
		host = "/var/run/postgresql/"
		port = 0
	}
	return provisioner.BuildConnectionString(dbname, user, password, host, port)
}

func init() {
	rootCmd.AddCommand(provisionCmd)

	provisionCmd.Flags().String("config", "", "Config file")
}
