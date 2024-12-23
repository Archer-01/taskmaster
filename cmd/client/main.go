package main

import (
	"fmt"

	"github.com/Archer-01/taskmaster/internal/utils"
)

func main() {
	message := utils.Hello("client")
	fmt.Println(message)

	for {
	}
}
