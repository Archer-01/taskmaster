package main

import (
	"fmt"
	"time"

	"github.com/Archer-01/taskmaster/internal/utils"

	"github.com/Archer-01/taskmaster/internal/parser/config"
)

func main() {
	message := utils.Hello("server")
	fmt.Println(message)

	_, err := config.Parse_config("taskmaster.tom")

	if (err != nil) {
		panic(err)
	}

	for {
		time.Sleep(8 * time.Second)
	}
}
