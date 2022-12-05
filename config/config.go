package config

import (
	"errors"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"os"
)

type Main struct {
	Database  string      `koanf:"database"`
	User      string      `koanf:"user"`
	Host      string      `koanf:"host"`
	Port      int         `koanf:"port"`
	Users     []*User     `koanf:"users"`
	Databases []*Database `koanf:"databases"`
}

type User struct {
	Name     string `koanf:"name"`
	Password string `koanf:"password"`
}

type Database struct {
	Name  string `koanf:"name"`
	Owner string `koanf:"owner"`
	User  string `koanf:"user"`
}

func Load(path string) (*Main, error) {
	k := koanf.New(".")

	err := k.Load(file.Provider(path), toml.Parser())

	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	var cfg Main
	err = k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "koanf"})
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (main *Main) GetUser(name string) *User {
	if main.Users == nil {
		return nil
	}
	for _, user := range main.Users {
		if user.Name == name {
			return user
		}
	}
	return nil
}
