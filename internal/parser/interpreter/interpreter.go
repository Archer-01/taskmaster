package interpreter

import (
	"fmt"
	"strings"
)

const (
	EXIT    = "exit"
	RELOAD  = "reload"
	RESTART = "restart"
	START   = "start"
	STATUS  = "status"
	STOP    = "stop"
	QUIT    = "quit"
)

func Parse(line string) []string {
	args := strings.Split(line, " ")

	switch args[0] {
	case "":
		return make([]string, 0)

	case EXIT:
		fmt.Print("\n")

	case RELOAD, RESTART, START, STATUS, STOP, QUIT:
		fmt.Printf("Running '%v'\n", args[0])

	default:
		fmt.Printf("*** Unknown syntax: %v\n", args[0])
	}
	return args
}
