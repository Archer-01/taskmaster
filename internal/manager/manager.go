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
	Done chan bool
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

	if conf.User != "" {
		fmt.Printf("[NOTICE] De-escalating privilege to user %v\n", conf.User)

		if err := utils.DeEscalatePrivilege(conf.User); err != nil {
			utils.Errorf(err.Error())
			os.Exit(1)
		}

		fmt.Println("[NOTICE] De-escalation successful")
	}

	jobs := make(map[string]*job.Job, 1)
	for name, prog := range conf.Programs {
		jobs[name] = job.NewJob(name, prog)
	}

	m.Jobs = jobs
	return nil
}

func (m *JobManager) start() {
	var done chan bool

	for _, j := range m.Jobs {
		if !j.Autostart {
			continue
		}

		done = make(chan bool, 1)
		defer close(done)
		j.Start(m.wg, done)
		<-done
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
			action.Done <- true
			return

		case RELOAD:
			utils.Logf("[RELOADING]")
			m.reload()
			action.Done <- true

		case START:
			utils.Logf("[STARTING] Program(name=%s)", action.Args[0])
			go m.Jobs[action.Args[0]].Start(m.wg, action.Done)

		case STOP:
			utils.Logf("[STOPPING] Program(name=%s)", action.Args[0])
			go m.Jobs[action.Args[0]].Stop(m.wg, action.Done)

		case RESTART:
			utils.Logf("[RESTARTING] Program(name=%s)", action.Args[0])
			go m.Jobs[action.Args[0]].Restart(m.wg, action.Done)
		}
	}
}

func (m *JobManager) stop() {
	var done chan bool

	for _, j := range m.Jobs {
		utils.Logf("[EXITING] Program(name=%s)", j.Name)
		done = make(chan bool, 1)
		defer close(done)
		j.Stop(m.wg, done)
		<-done
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
