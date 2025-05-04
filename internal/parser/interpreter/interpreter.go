package interpreter

import (
	"fmt"
	"strings"
)

const (
	RELOAD  = "reload"
	RESTART = "restart"
	START   = "start"
	STATUS  = "status"
	STOP    = "stop"
	QUIT    = "quit"
)

func Parse(line string) ([]string, error) {
	args := strings.Split(line, " ")

	if len(args) == 0 {
		return make([]string, 0), nil
	}

	switch args[0] {
	case RELOAD, RESTART, START, STATUS, STOP, QUIT:
		return args, nil

	default:
		return nil, fmt.Errorf("*** Unknown syntax: %v", args[0])
	}
}
