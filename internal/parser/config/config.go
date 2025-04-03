package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

type Program struct {
	Command   string `toml:"command" validate:"required"`
	Autostart bool   `toml:"autostart" validate:"required"`
	// Autorestart   string `toml:"autorestart"`
	// NumProcs      int    `toml:"numprocs"`
	// StdoutLogFile string `toml:"stdout_logfile`
	// StderrLogFile string `toml:"stderr_logfile"`
}

type Config struct {
	Programs map[string]Program `toml:"program"`
}

func validate(val interface{}) error {
	v := reflect.ValueOf(val)
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := v.Type().Field(i).Name
		tag := v.Type().Field(i).Tag.Get("validate")
		if tag != "" {
			rules := strings.Split(tag, ",")
			for _, rule := range rules {
				switch {
				case rule == "required":
					if field.String() == "" {
						return fmt.Errorf(
							"%s is required",
							fieldName,
						)
					}
				}
			}
		}
		switch {
		case field.Kind() == reflect.Struct:
			return validate(field.Interface())
		case field.Kind() == reflect.Map:
			for _, key := range field.MapKeys() {
				value := field.MapIndex(key).Interface()
				if reflect.ValueOf(value).Kind() == reflect.Struct {
					err_msg := validate(value)
					if err_msg != nil {
						return err_msg
					}
				}
			}
		}
	}
	return nil
}

func Parse_config(file string) (Config, error) {
	var conf Config

	_, err := toml.DecodeFile(file, &conf)

	err_msg := validate(conf)
	if err_msg != nil {
		return conf, err_msg
	}

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
