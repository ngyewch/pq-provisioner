# pq-provisioner

Provision PostgreSQL databases and users.

## Usage

```
pq-provisioner provision --config (config file)
```

## Config file

```
database = "postgres"  # Admin user database. [OPTIONAL]
user = "postgres"      # Admin user. [REQUIRED]
host = "localhost"     # Server host to connect to. If admin user password is not specified, defaults to UNIX domain socket "/var/run/postgresql/". [OPTIONAL] 
port = 5432            # Server port to connect to. [OPTIONAL]
sslmode = "disable"    # SSL mode. [OPTIONAL]
sshProxy = "alias"     # SSH proxy alias. [OPTIONAL]

[[users]]
name = "postgres"      # User name. [REQUIRED]
password = "password"  # User password. [OPTIONAL]

[[users]]
name = "app_admin"
password = "app_admin_password"

[[users]]
name = "app_user"
password = "app_user_password"

[[databases]]
name = "test"          # Database name. [REQUIRED]
owner = "app_admin"    # Database owner. [REQUIRED]
users = ["app_user"]   # Database users. [OPTIONAL]
```
