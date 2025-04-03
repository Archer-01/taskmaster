package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Archer-01/taskmaster/internal/utils"

	"github.com/Archer-01/taskmaster/internal/parser/config"
)

func main() {
	message := utils.Hello("server")
	fmt.Println(message)

	_, err := config.Parse_config("taskmaster.toml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal: %s\n", err)
		os.Exit(1)
	}

	for {
		time.Sleep(8 * time.Second)
	}
}
