package config

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
)

type Program struct {
	Command       string `toml:"command" validate:"required"`
	Autostart     bool   `toml:"autostart" validate:"default=true"`
	NumProcs      int    `toml:"numprocs" validate:"default=1,min=1"`
	StdoutLogFile string `toml:"stdout_logfile"`
	StderrLogFile string `toml:"stderr_logfile"`
}

type Config struct {
	Programs map[string]*Program `toml:"program"`
}

func ParseCommand(cmd string) []string {
	return strings.FieldsFunc(cmd, func(r rune) bool {
		return strings.ContainsRune(" \t\n\v\f\r", r)
	})
}

func ParseConfig(file string) (Config, error) {
	var conf Config

	mdata, err := toml.DecodeFile(file, &conf)

	err_msg := Validate(&conf, mdata)
	if err_msg != nil {
		return conf, err_msg
	}

	for name, program := range conf.Programs {
		fmt.Printf("Program: %s\n", name)
		fmt.Printf("\tCommand: %s\n", program.Command)
		fmt.Printf("\tAutostart: %s\n", program.Autostart)
		// fmt.Printf("\tAutorestart: %s\n", program.Autorestart)
		fmt.Printf("\tNumProcs: %s\n", program.NumProcs)
		// fmt.Printf("\tStdoutLogFile: %s\n", program.StdoutLogFile)
		// fmt.Printf("\tStderrLogFile: %s\n", program.StderrLogFile)
	}

	return conf, err
}
