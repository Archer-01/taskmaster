package main

import (
	"io"
	"strings"

	"github.com/Archer-01/taskmaster/internal/client"
	"github.com/Archer-01/taskmaster/internal/parser/interpreter"
	"github.com/Archer-01/taskmaster/internal/utils"
	"github.com/chzyer/readline"
)

func main() {
	setup, err := utils.ParseSetupFile()
	if err != nil {
		utils.Errorf(err.Error())
		return
	}

	client, err := client.NewClient(setup.Socket)
	if err != nil {
		utils.Errorf(err.Error())
		return
	}
	defer client.Close()

	rl, err := readline.New(setup.Prompt)
	if err != nil {
		panic(err)
	}

	defer rl.Close()
	rl.Config.EOFPrompt = ""

	for {
		line, err := rl.Readline()
		if err == io.EOF {
			return
		}
		if err != nil {
			break
		}

		if line == "" {
			continue
		}

		args := interpreter.Parse(line)
		if args[0] == interpreter.EXIT {
			return
		}

		err = client.Send(strings.Join(args, " "))
		if err != nil {
			utils.Errorf("%s", err.Error())
		}
	}
}
