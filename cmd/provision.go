package cmd

import (
	"database/sql"
	"fmt"
	"github.com/spf13/cobra"
	"strings"

	_ "github.com/lib/pq"
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
	config, err := loadConfig(cmd)
	if err != nil {
		return err
	}

	connStrParts := make([]string, 0)
	{
		if config.Database != "" {
			connStrParts = append(connStrParts, fmt.Sprintf("dbname=%s", config.Database))
		}
		user := config.GetUser(config.User)
		if config.User != "" {
			connStrParts = append(connStrParts, fmt.Sprintf("user=%s", config.User))
			if (user != nil) && (user.Password != "") {
				connStrParts = append(connStrParts, fmt.Sprintf("password=%s", user.Password))
			}
		}
		if config.Host != "" {
			if (config.Host == "localhost") && (user.Password == "") {
				connStrParts = append(connStrParts, "host=/var/run/postgresql/")

			} else {
				connStrParts = append(connStrParts, fmt.Sprintf("host=%s", config.Host))
			}
		}
		if config.Port != 0 {
			connStrParts = append(connStrParts, fmt.Sprintf("port=%d", config.Port))
		}
	}

	connStr := strings.Join(connStrParts, " ")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	provisioner, err := NewProvisioner(db)
	if err != nil {
		return err
	}

	for _, database := range config.Databases {
		if !provisioner.HasDatabase(database.Name) {
			log.Infof("Creating database %s...", database.Name)
			err := provisioner.CreateDatabase(database.Name)
			if err != nil {
				return err
			}
		}
		if !provisioner.HasUser(database.Owner) {
			log.Infof("Creating user %s...", database.Owner)
			user := config.GetUser(database.Owner)
			if user == nil {
				return fmt.Errorf("user not defined")
			}
			if user.Password == "" {
				return fmt.Errorf("user password not specified")
			}
			err := provisioner.CreateUser(user.Name, user.Password)
			if err != nil {
				return err
			}
		}
		if !provisioner.HasUser(database.User) {
			log.Infof("Creating user %s...", database.User)
			user := config.GetUser(database.User)
			if user == nil {
				return fmt.Errorf("user not defined")
			}
			if user.Password == "" {
				return fmt.Errorf("user password not specified")
			}
			err := provisioner.CreateUser(user.Name, user.Password)
			if err != nil {
				return err
			}
		}
		log.Infof("Setting database %s owner to %s...", database.Name, database.Owner)
		err = provisioner.SetDatabaseOwner(database.Name, database.Owner)
		if err != nil {
			return err
		}
		{
			connStrParts := make([]string, 0)
			if database.Name != "" {
				connStrParts = append(connStrParts, fmt.Sprintf("dbname=%s", database.Name))
			}
			connStrParts = append(connStrParts, fmt.Sprintf("user=%s", database.Owner))
			user := config.GetUser(database.Owner)
			if (user != nil) && (user.Password != "") {
				connStrParts = append(connStrParts, fmt.Sprintf("password=%s", user.Password))
			}
			if config.Host != "" {
				connStrParts = append(connStrParts, fmt.Sprintf("host=%s", config.Host))
			}
			if config.Port != 0 {
				connStrParts = append(connStrParts, fmt.Sprintf("port=%d", config.Port))
			}

			connStr = strings.Join(connStrParts, " ")
			db2, err := sql.Open("postgres", connStr)
			if err != nil {
				return err
			}

			log.Infof("Setting database %s user to %s...", database.Name, database.User)
			err = setDatabaseUser(db2, database.User)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type Provisioner struct {
	db            *sql.DB
	usernames     []string
	databaseNames []string
}

func NewProvisioner(db *sql.DB) (*Provisioner, error) {
	p := &Provisioner{
		db: db,
	}
	usernames, err := p.getUsernames()
	if err != nil {
		return nil, err
	}
	p.usernames = usernames
	databaseNames, err := p.getDatabaseNames()
	if err != nil {
		return nil, err
	}
	p.databaseNames = databaseNames
	return p, nil
}

func (p *Provisioner) getUsernames() ([]string, error) {
	rows, err := p.db.Query("SELECT usename FROM pg_catalog.pg_user")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	usernames := make([]string, 0)
	for {
		if !rows.Next() {
			break
		}
		var username string
		err = rows.Scan(&username)
		if err != nil {
			return nil, err
		}
		usernames = append(usernames, username)
	}
	return usernames, nil
}

func (p *Provisioner) getDatabaseNames() ([]string, error) {
	rows, err := p.db.Query("SELECT datname FROM pg_catalog.pg_database")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	databaseNames := make([]string, 0)
	for {
		if !rows.Next() {
			break
		}
		var databaseName string
		err = rows.Scan(&databaseName)
		if err != nil {
			return nil, err
		}
		databaseNames = append(databaseNames, databaseName)
	}
	return databaseNames, nil
}

func (p *Provisioner) HasDatabase(name string) bool {
	return stringArrayContains(p.databaseNames, name)
}

func (p *Provisioner) HasUser(name string) bool {
	return stringArrayContains(p.usernames, name)
}

func (p *Provisioner) CreateDatabase(name string) error {
	_, err := p.db.Exec(fmt.Sprintf("CREATE DATABASE %s", name))
	if err != nil {
		return err
	}
	p.databaseNames = append(p.databaseNames, name)
	return nil
}

func (p *Provisioner) CreateUser(name string, password string) error {
	_, err := p.db.Exec(fmt.Sprintf("CREATE USER %s", name))
	if err != nil {
		return err
	}
	if password != "" {
		_, err = p.db.Exec(fmt.Sprintf("ALTER USER %s WITH PASSWORD '%s'", name, password))
		if err != nil {
			return err
		}
	}
	p.usernames = append(p.usernames, name)
	return nil
}

func (p *Provisioner) SetDatabaseOwner(databaseName string, userName string) error {
	_, err := p.db.Exec(fmt.Sprintf("ALTER DATABASE %s OWNER TO %s", databaseName, userName))
	if err != nil {
		return err
	}
	return nil
}

func setDatabaseUser(db *sql.DB, userName string) error {
	_, err := db.Exec(fmt.Sprintf("GRANT USAGE ON SCHEMA public TO %s", userName))
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("GRANT SELECT, UPDATE, INSERT, DELETE ON ALL TABLES IN SCHEMA public TO %s", userName))
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO %s", userName))
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, UPDATE, INSERT, DELETE ON TABLES TO %s", userName))
	if err != nil {
		return err
	}
	_, err = db.Exec(fmt.Sprintf("ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO %s", userName))
	if err != nil {
		return err
	}
	return nil
}

func stringArrayContains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(provisionCmd)

	provisionCmd.Flags().String("config", "", "Config file")
}
