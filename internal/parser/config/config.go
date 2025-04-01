package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Program struct {
	Command       string `toml:"command"`
	Autostart     bool   `toml:"autostart"`
	// Autorestart   string `toml:"autorestart"`
	// NumProcs      int    `toml:"numprocs"`
	// StdoutLogFile string `toml:"stdout_logfile`
	// StderrLogFile string `toml:"stderr_logfile"`
}

type Config struct {
	Programs map[string]Program `toml:"program"`
}

func Parse_config(file string) (Config, error) {
	var conf Config

	_, err := toml.DecodeFile(file, &conf)

	for name, program := range conf.Programs {
		fmt.Printf("Program: %s\n", name)
		fmt.Printf("\tCommand: %s\n", program.Command)
		// fmt.Printf("\tAutostart: %s\n", program.Autostart)
		// fmt.Printf("\tAutorestart: %s\n", program.Autorestart)
		// fmt.Printf("\tNumProcs: %s\n", program.NumProcs)
		// fmt.Printf("\tStdoutLogFile: %s\n", program.StdoutLogFile)
		// fmt.Printf("\tStderrLogFile: %s\n", program.StderrLogFile)
	}

	return conf, err
}
