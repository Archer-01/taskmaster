package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
}

func Parse_config(file string) (Config, error) {
	var conf Config

	_, err := toml.DecodeFile(file, &conf)

	return conf, err
}
