package provisioner

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"strings"
)

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

func SetDatabaseUser(db *sql.DB, userName string) error {
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

func BuildConnectionString(dbname string, user string, password string, host string, port int) string {
	connStrParts := make([]string, 0)
	if dbname != "" {
		connStrParts = append(connStrParts, fmt.Sprintf("dbname=%s", dbname))
	}
	if user != "" {
		connStrParts = append(connStrParts, fmt.Sprintf("user=%s", user))
	}
	if password != "" {
		connStrParts = append(connStrParts, fmt.Sprintf("password=%s", password))
	}
	if host != "" {
		connStrParts = append(connStrParts, fmt.Sprintf("host=%s", host))
	}
	if (port != 0) && (port != 5432) {
		connStrParts = append(connStrParts, fmt.Sprintf("port=%d", port))
	}
	return strings.Join(connStrParts, " ")
}
