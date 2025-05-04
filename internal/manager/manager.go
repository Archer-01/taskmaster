package manager

import (
	"fmt"
	"os"
	"sync"

	"github.com/Archer-01/taskmaster/internal/job"
	"github.com/Archer-01/taskmaster/internal/parser/config"
	"github.com/Archer-01/taskmaster/internal/utils"
)

const (
	QUIT    = "quit"
	RELOAD  = "reload"
	START   = "start"
	STOP    = "stop"
	RESTART = "restart"
)

type Action struct {
	Type string
	Args []string
}

type JobManager struct {
	Jobs    map[string]*job.Job
	Config  string
	actions chan Action
	sigs    chan os.Signal
	wg      *sync.WaitGroup
}

func NewJobManager(path string, wg *sync.WaitGroup) *JobManager {
	return &JobManager{
		Config:  path,
		actions: make(chan Action, 1),
		wg:      wg,
	}
}

func (m *JobManager) Init() error {
	conf, err := config.ParseConfig(m.Config)
	if err != nil {
		return err
	}

	jobs := make(map[string]*job.Job, 1)
	for name, prog := range conf.Programs {
		jobs[name] = job.NewJob(name, prog)
	}

	m.Jobs = jobs
	return nil
}

func (m *JobManager) start() {
	for _, j := range m.Jobs {
		if !j.Autostart {
			continue
		}

		j.Start(m.wg)
	}
}

func (m *JobManager) Run() {
	m.start()
	for {
		action := <-m.actions
		switch action.Type {

		case QUIT:
			m.stop()
			utils.Logf("[QUITTING]")
			m.finish()
			return

		case RELOAD:
			utils.Logf("[RELOADING]")
			m.reload()

		case START:
			utils.Logf("[STARTING] Program(name=%s)", action.Args[0])
			m.Jobs[action.Args[0]].Start(m.wg)

		case STOP:
			utils.Logf("[STOPPING] Program(name=%s)", action.Args[0])
			m.Jobs[action.Args[0]].Stop()

		case RESTART:
			utils.Logf("[RESTARTING] Program(name=%s)", action.Args[0])
			m.Jobs[action.Args[0]].Restart(m.wg)
		}
	}
}

func (m *JobManager) stop() {
	for _, j := range m.Jobs {
		utils.Logf("[EXITING] Program(name=%s)", j.Name)

		j.Stop()
		j.WaitJob()
	}
}

func (m *JobManager) finish() {
	close(m.actions)
}

func (m *JobManager) reload() {
	m.stop()

	m.Init()
	fmt.Println(os.Getpid())

	m.start()
}
