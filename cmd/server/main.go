package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/Archer-01/taskmaster/internal/manager"
	"github.com/Archer-01/taskmaster/internal/server"
	"github.com/Archer-01/taskmaster/internal/utils"
)

func main() {
	message := utils.Hello("server")
	fmt.Println(message)

	setup, err := utils.ParseSetupFile()
	if err != nil {
		log.Fatal(err)
	}

	Manager := manager.NewJobManager(setup.Config)
	err = Manager.Init()
	if err != nil {
		log.Fatal(err)
	}

	Server := server.NewServer(setup.Socket, Manager)
	err = Server.Init()
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	Manager.InitSignals()
	go Manager.WaitForSignals(&wg)
	defer Manager.StopSignals()

	go Server.Start(&wg)
	defer Server.Stop()

	Manager.Run()
}
