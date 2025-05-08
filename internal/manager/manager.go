package manager

import (
	"fmt"
	"os"

	"github.com/Archer-01/taskmaster/internal/job"
	"github.com/Archer-01/taskmaster/internal/parser/config"
	"github.com/Archer-01/taskmaster/internal/utils"
)

const (
	QUIT   = "quit"
	RELOAD = "reload"
)

type Action struct {
	t string
}

type JobManager struct {
	Jobs    []*job.Job
	Config  string
	actions chan string
	sigs    chan os.Signal
}

func NewJobManager(path string) *JobManager {
	return &JobManager{
		Config:  path,
		actions: make(chan string, 1),
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

	var jobs []*job.Job
	for name, prog := range conf.Programs {
		jobs = append(jobs, job.NewJob(name, prog))
	}

	m.Jobs = jobs
	return nil
}

func (m *JobManager) start() {
	for _, j := range m.Jobs {
		if !j.Autostart {
			continue
		}

		err := j.StartJob()
		if err != nil {
			utils.Errorf(err.Error())
		}
	}
}

func (m *JobManager) Execute(action string, args ...string) error {
	switch action {
	case QUIT:
		m.actions <- QUIT
	case RELOAD:
		m.actions <- RELOAD
	default:
		return fmt.Errorf("%s Unknown command", action)
	}
	return nil
}

func (m *JobManager) Run() {
	m.start()
	for {
		action := <-m.actions
		switch action {

		case QUIT:
			m.stop()
			m.finish()
			return

		case RELOAD:
			m.reload()

		}
	}
}

func (m *JobManager) stop() {
	for _, j := range m.Jobs {
		// HACK: This is a temporary fix
		// Jobs should have a state (Started, Stopped, Running, ...etc)
		// and that status must be checked before trying to finish a job
		if !j.Autostart {
			continue
		}

		utils.Logf("Exiting [%s]", j.Name)

		err := j.Stop()
		if err != nil {
			utils.Errorf(err.Error())
		}
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
