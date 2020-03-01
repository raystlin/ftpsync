package config

import (
	"encoding/json"
	"os"

	"github.com/urfave/cli"
)

type ServerConfig struct {
	Server   string `json:"server"`
	Username string `json:"username"`
	Password string `json:"password"`
	Path     string `json:"path"`
}

type Config struct {
	Source     ServerConfig `json:"source"`
	Dest       ServerConfig `json:"dest"`
	NumClients int          `json:"num-clients"`
}

func ReadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	conf := new(Config)
	dec := json.NewDecoder(f)

	if conf.NumClients < 1 {
		conf.NumClients = 1
	}

	return conf, dec.Decode(conf)
}

func FromContext(c *cli.Context) *Config {
	return c.App.Metadata["config"].(*Config)
}
