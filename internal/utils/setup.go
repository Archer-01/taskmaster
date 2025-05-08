package utils

import (
	"strconv"
	"syscall"
	"os/user"

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

func DeEscalatePrivilege(username string) error {
	u, err := user.Lookup(username)

	if err != nil {
		return err
	}

	newUid, err := strconv.Atoi(u.Uid)

	if err != nil {
		return err
	}

	if err := syscall.Setuid(newUid); err != nil {
		return err
	}

	return nil
}
