package provisioner

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/ngyewch/go-pqssh"
	ssh_helper "github.com/ngyewch/go-ssh-helper"
	"github.com/ngyewch/pq-provisioner/config"
	"golang.org/x/crypto/ssh"
)

var (
	log = slog.Default().With("pkg", "provisioner")
)

type ConfigProvisioner struct {
	cfg       *config.Main
	sshClient *ssh.Client
}

func NewConfigProvisioner(cfg *config.Main, sshClientFactory *ssh_helper.SSHClientFactory) (*ConfigProvisioner, error) {
	p := &ConfigProvisioner{
		cfg: cfg,
	}
	if cfg.SshProxy != "" {
		log.LogAttrs(context.Background(), slog.LevelInfo, "Creating ssh proxy",
			slog.String("proxy", cfg.SshProxy),
		)
		if sshClientFactory == nil {
			sshClientFactory = ssh_helper.DefaultSSHClientFactory()
		}
		sshClient, err := sshClientFactory.CreateForAlias(cfg.SshProxy)
		if err != nil {
			return nil, err
		}
		p.sshClient = sshClient
	}
	return p, nil
}

func (p *ConfigProvisioner) Close() error {
	if p.sshClient != nil {
		err := p.sshClient.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *ConfigProvisioner) Provision() error {
	db, err := p.openDB(p.cfg.Database, p.cfg.User)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	prov, err := NewProvisioner(db)
	if err != nil {
		return err
	}

	for _, database := range p.cfg.Databases {
		err = p.createDatabaseIfNotExist(prov, database.Name)
		if err != nil {
			return err
		}

		err = p.createUserIfNotExist(prov, database.Owner)
		if err != nil {
			return err
		}

		for _, user := range database.Users {
			err = p.createUserIfNotExist(prov, user)
			if err != nil {
				return err
			}
		}

		log.LogAttrs(context.Background(), slog.LevelInfo, "Setting database owner",
			slog.String("dbname", database.Name),
			slog.String("user", database.Owner),
		)
		err = prov.SetDatabaseOwner(database.Name, database.Owner)
		if err != nil {
			return err
		}

		err = func() error {
			db2, err := p.openDB(database.Name, database.Owner)
			if err != nil {
				return err
			}
			defer func(db *sql.DB) {
				_ = db.Close()
			}(db2)

			for _, user := range database.Users {
				log.LogAttrs(context.Background(), slog.LevelInfo, "Setting database user",
					slog.String("dbname", database.Name),
					slog.String("user", user),
				)
				err = SetDatabaseUser(db2, user)
				if err != nil {
					return err
				}
			}

			return nil
		}()
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *ConfigProvisioner) createDatabaseIfNotExist(prov *Provisioner, dbname string) error {
	if prov.HasDatabase(dbname) {
		return nil
	}

	log.LogAttrs(context.Background(), slog.LevelInfo, "Creating database",
		slog.String("dbname", dbname),
	)

	err := prov.CreateDatabase(dbname)
	if err != nil {
		return err
	}

	return nil
}

func (p *ConfigProvisioner) createUserIfNotExist(prov *Provisioner, username string) error {
	if prov.HasUser(username) {
		return nil
	}

	log.LogAttrs(context.Background(), slog.LevelInfo, "Creating user",
		slog.String("user", username),
	)

	user := p.cfg.GetUser(username)
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

	return nil
}

func (p *ConfigProvisioner) openDB(dbname string, user string) (*sql.DB, error) {
	dsn := p.buildConnectionString(dbname, user, p.cfg.Host, p.cfg.Port, p.cfg.SslMode)
	if p.sshClient != nil {
		dbConnector := pqssh.NewConnector(p.sshClient, dsn)
		return sql.OpenDB(dbConnector), nil
	} else {
		return sql.Open("postgres", dsn)
	}
}

func (p *ConfigProvisioner) buildConnectionString(dbname string, user string, host string, port int, sslmode string) string {
	userEntry := p.cfg.GetUser(user)
	password := ""
	if userEntry != nil {
		password = userEntry.Password
	}
	if (host == "") && (password == "") {
		host = "/var/run/postgresql/"
		port = 0
	}
	return BuildConnectionString(dbname, user, password, host, port, sslmode)
}
