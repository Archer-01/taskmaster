package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Archer-01/taskmaster/internal/manager"
	"github.com/Archer-01/taskmaster/internal/utils"

	"github.com/Archer-01/taskmaster/internal/parser/config"
)

func main() {
	message := utils.Hello("server")
	fmt.Println(message)

	conf, err := config.Parse_config("taskmaster.toml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal: %s\n", err)
		os.Exit(1)
	}

	ins := manager.Init(conf)

	ins.Start()

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGQUIT)

	done := make(chan bool, 1)

	go func() {
		sig := <-sigs
		fmt.Fprintf(os.Stdout, "\nCaught signal: '%s'\n", sig)
		ins.Finish()
		done <- true
	}()

	fmt.Println("Waiting for Ctrl-C")
	<-done
	fmt.Println("exiting")
}
