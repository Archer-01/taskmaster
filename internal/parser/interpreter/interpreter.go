package interpreter

import (
	"fmt"
	"os"
	"strings"
)

func Parse(line string) {
	args := strings.Split(line, " ")

	switch args[0] {
	case "":
		return

	case "exit":
		fmt.Print("\n")
		os.Exit(0)

	case "reload", "restart", "start", "status", "stop":
		fmt.Printf("Running '%v'\n", args[0])

	default:
		fmt.Printf("*** Unknown syntax: %v\n", args[0])
		os.Exit(1)
	}
}
