package utils

import (
	"github.com/Archer-01/taskmaster/internal/parser/config"
	"github.com/BurntSushi/toml"
)

type Setup struct {
	Prompt string `toml:"prompt"`
	Socket string `toml:"socket" validate:"default=/tmp/taskmaster.sock"`
	Config string `toml:"config" validate:"default=taskmaster.toml"`
}

const (
	CONF = "setup.toml"
)

func ParseSetupFile() (Setup, error) {
	var setup Setup

	mdata, err := toml.DecodeFile(CONF, &setup)
	if err != nil {
		return setup, err
	}

	err = config.Validate(&setup, mdata)
	if err != nil {
		return setup, err
	}

	return setup, nil
}
