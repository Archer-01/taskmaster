package manager

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Archer-01/taskmaster/internal/utils"
)

func (m *JobManager) InitSignals() {
	m.sigs = make(chan os.Signal, 1)
	signal.Notify(m.sigs, syscall.SIGQUIT, syscall.SIGHUP)
}

func (m *JobManager) StopSignals() {
	close(m.sigs)
}

func (m *JobManager) WaitForSignals(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		sig, ok := <-m.sigs
		if !ok {
			return
		}

		utils.Logf("\nCaught signal: '%s'\n", sig)

		switch sig {

		case syscall.SIGQUIT:
			m.actions <- QUIT
			return

		case syscall.SIGHUP:
			m.actions <- RELOAD

		}
	}
}
