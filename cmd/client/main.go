package main

import (
	"github.com/Archer-01/taskmaster/internal/parser/interpreter"
	"github.com/chzyer/readline"
)

const prompt string = "taskmaster> "

func main() {
	rl, err := readline.New(prompt)

	if err != nil {
		panic(err)
	}

	defer rl.Close()
	rl.Config.EOFPrompt = ""

	for {
		line, err := rl.Readline()

		if err != nil {
			break
		}

		interpreter.Parse(line)
	}
}
