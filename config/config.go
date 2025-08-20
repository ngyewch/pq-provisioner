package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/mitchellh/mapstructure"
)

type Main struct {
	Database  string      `koanf:"database"`
	User      string      `koanf:"user" validate:"required"`
	Host      string      `koanf:"host"`
	Port      int         `koanf:"port"`
	SslMode   string      `koanf:"sslmode"`
	SshProxy  string      `koanf:"sshProxy"`
	Users     []*User     `koanf:"users" validate:"dive"`
	Databases []*Database `koanf:"databases" validate:"dive"`
}

type User struct {
	Name     string `koanf:"name" validate:"required"`
	Password string `koanf:"password"`
}

type Database struct {
	Name  string   `koanf:"name" validate:"required"`
	Owner string   `koanf:"owner" validate:"required"`
	Users []string `koanf:"users"`
}

func Load(provider koanf.Provider, parser koanf.Parser) (*Main, error) {
	k := koanf.New(".")

	err := k.Load(provider, parser)

	if (err != nil) && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	var cfg Main
	err = k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{
		Tag: "koanf",
		DecoderConfig: &mapstructure.DecoderConfig{
			Result:      &cfg,
			ErrorUnused: true,
			ErrorUnset:  false,
		},
	})
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	err = validate.Struct(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func LoadFromFile(path string) (*Main, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return nil, os.ErrNotExist
	}
	parser, err := getParserFromPath(path)
	if err != nil {
		return nil, err
	}
	return Load(file.Provider(path), parser)
}

func getParserFromPath(path string) (koanf.Parser, error) {
	ext := filepath.Ext(path)
	switch ext {
	case ".toml":
		return toml.Parser(), nil
	case ".yaml":
		return yaml.Parser(), nil
	case ".yml":
		return yaml.Parser(), nil
	case ".json":
		return json.Parser(), nil
	}
	return nil, fmt.Errorf("unsupported file extension")
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
