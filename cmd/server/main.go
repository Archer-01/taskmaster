package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Archer-01/taskmaster/internal/manager"
	"github.com/Archer-01/taskmaster/internal/parser/config"
	"github.com/Archer-01/taskmaster/internal/utils"
)

func main() {
	message := utils.Hello("server")
	fmt.Println(message)

	var ins *manager.JobManager

	reload := make(chan os.Signal, 1)
	signal.Notify(reload, syscall.SIGHUP)

	go func() {
		for {
			conf, err := config.Parse_config("taskmaster.toml")
			if err != nil {
				log.Fatalln("Fatal:", err)
			}

			ins = manager.Init(conf)
			fmt.Println(os.Getpid())

			ins.Start()

			<-reload
			ins.Finish()
		}
	}()

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
