package manager

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Archer-01/taskmaster/internal/logger"
)

func (m *JobManager) InitSignals() {
	m.sigs = make(chan os.Signal, 1)
	signal.Notify(m.sigs, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
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

		logger.Warnf("Caught signal: %s", sig)

		switch sig {

		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			m.actions <- Action{Type: QUIT}
			return

		case syscall.SIGHUP:
			m.actions <- Action{Type: RELOAD}

		}
	}
}
