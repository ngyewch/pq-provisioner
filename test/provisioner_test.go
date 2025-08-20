package test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	ssh_helper "github.com/ngyewch/go-ssh-helper"
	"github.com/ngyewch/pq-provisioner/config"
	"github.com/ngyewch/pq-provisioner/provisioner"
	"github.com/trzsz/ssh_config"
)

func Test1(t *testing.T) {
	env, err := NewEnv(t)
	if err != nil {
		t.Fatal(err)
	}
	defer func(env *Env) {
		_ = env.Close()
	}(env)

	cfg, err := config.LoadFromFile(filepath.Join("resources", "config", "test1.toml"))
	if err != nil {
		t.Fatal(err)
	}

	postgresFactory, err := NewPostgresFactory(env.Helper())
	if err != nil {
		t.Fatal(err)
	}

	postgresContainer, err := postgresFactory.Start("pg1", "password", 15432)
	if err != nil {
		t.Fatal(err)
	}

	err = env.Helper().ConnectNetworks(postgresContainer, env.Network())
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Second)

	configProvisioner, err := provisioner.NewConfigProvisioner(cfg, nil)
	if err != nil {
		t.Fatal(err)
	}

	err = configProvisioner.Provision()
	if err != nil {
		t.Fatal(err)
	}
}

func Test2(t *testing.T) {
	env, err := NewEnv(t)
	if err != nil {
		t.Fatal(err)
	}
	defer func(env *Env) {
		_ = env.Close()
	}(env)

	cfg, err := config.LoadFromFile(filepath.Join("resources", "config", "test2.toml"))
	if err != nil {
		t.Fatal(err)
	}

	sshServerFactory, err := NewOpenSSHServerFactory(env.Helper())
	if err != nil {
		t.Fatal(err)
	}

	postgresFactory, err := NewPostgresFactory(env.Helper())
	if err != nil {
		t.Fatal(err)
	}

	err = generateSshKeyPair("/tmp/pq-provisioner-test/test2", "test2")
	if err != nil {
		t.Fatal(err)
	}
	publicKeyBytes, err := os.ReadFile("/tmp/pq-provisioner-test/test2/test2.pub")
	if err != nil {
		t.Fatal(err)
	}
	publicKey := string(publicKeyBytes)

	sshServerContainer, err := sshServerFactory.StartSshHost("proxy", "bob", publicKey, 10022)
	if err != nil {
		t.Fatal(err)
	}

	postgresContainer, err := postgresFactory.Start("pg1", "password", 0)
	if err != nil {
		t.Fatal(err)
	}

	err = env.Helper().ConnectNetworks(sshServerContainer, env.Network())
	if err != nil {
		t.Fatal(err)
	}

	err = env.Helper().ConnectNetworks(postgresContainer, env.Network())
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(10 * time.Second)

	userSettings := &ssh_config.UserSettings{}
	userSettings.ConfigFinder(func() string {
		return filepath.Join("resources", "ssh_config", "test2")
	})
	sshClientFactory := ssh_helper.NewSSHClientFactory(userSettings)

	configProvisioner, err := provisioner.NewConfigProvisioner(cfg, sshClientFactory)
	if err != nil {
		t.Fatal(err)
	}

	err = configProvisioner.Provision()
	if err != nil {
		t.Fatal(err)
	}
}
