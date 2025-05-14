package manager

import (
	"os"
	"sync"

	"github.com/Archer-01/taskmaster/internal/job"
	"github.com/Archer-01/taskmaster/internal/logger"
	"github.com/Archer-01/taskmaster/internal/parser/config"
	"github.com/Archer-01/taskmaster/internal/utils"
)

const (
	QUIT    = "quit"
	RELOAD  = "reload"
	START   = "start"
	STOP    = "stop"
	RESTART = "restart"
	ALL     = "all"
)

type Action struct {
	Type string
	Args []string
	Data chan string
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
		logger.Infof("De-escalating privilege to user %s", conf.User)

		if err := utils.DeEscalatePrivilege(conf.User); err != nil {
			logger.Critical(err)
			os.Exit(1)
		}

		logger.Info("Privilege de-escalation successful")
	}

	logger.Infof("taskmasterd started with pid %d", os.Getpid())

	jobs := make(map[string]*job.Job, 1)
	for name, prog := range conf.Programs {
		if name == ALL {
			return fmt.Errorf("all is a special name, please use another name")
		}
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
			logger.Info("Quitting...")
			m.finish()
			action.Done <- true
			return

		case RELOAD:
			logger.Warn("Reloading...")
			m.reload()
			action.Done <- true

		case START:
			m.setJobs("STARTING", (*job.Job).Start, action)

		case STOP:
			m.setJobs("STOPPING", (*job.Job).Stop, action)

		case RESTART:
			m.setJobs("RESTARTING", (*job.Job).Restart, action)

		case STATUS:
			m.getStatus(action)

		default:
			action.Data <- "unknown command " + action.Type
			action.Done <- false
		}
	}
}

func (m *JobManager) runWorkerJob(j *job.Job, worker job.WorkerFn, done chan bool, state string) {
	utils.Logf("[%s] Program(name=%s)", state, j.Name)
	go worker(j, m.wg, done)
}

func (m *JobManager) runWorkerJobs(jobs map[string]*job.Job, worker job.WorkerFn, action Action, state string) {
	jobs_done := []chan bool{}
	for _, j := range jobs {
		_done := make(chan bool, 1)
		jobs_done = append(jobs_done, _done)
		m.runWorkerJob(j, worker, _done, state)
	}
	for _, _done := range jobs_done {
		defer close(_done)
		<-_done
	}
	action.Done <- true
}

func (m *JobManager) setJobs(state string, worker job.WorkerFn, action Action) {
	if len(action.Args) != 1 {
		action.Data <- "command accepts 1 argument only"
		action.Done <- false
		return
	}
	name := action.Args[0]
	if name != ALL {
		j, found := m.Jobs[name]
		if !found {
			action.Data <- "job is not recognized"
			action.Done <- false
			return
		}
		m.runWorkerJob(j, worker, action.Done, state)
	} else {
		m.runWorkerJobs(m.Jobs, worker, action, state)
	}
}

func (m *JobManager) getStatus(action Action) {
	if len(action.Args) != 1 {
		action.Data <- "command accepts 1 argument only"
		action.Done <- false
		return
	}
	getStatusFmt := func(j *job.Job) string { return "[" + j.Name + "]: " + j.State }

	name := action.Args[0]
	if name != ALL {
		j, found := m.Jobs[name]
		if !found {
			action.Data <- "job is not recognized"
			action.Done <- false
			return
		}
		action.Data <- getStatusFmt(j)
		action.Done <- true

	} else {
		msg := ""
		i := 0
		for _, j := range m.Jobs {
			msg += getStatusFmt(j)
			i++
			if i != len(m.Jobs) {
				msg += "\n"
			}
		}

		action.Data <- msg
		action.Done <- true
	}
}

func (m *JobManager) stop() {
	var done chan bool

	for _, j := range m.Jobs {
		logger.Infof("Exiting program %s", j.Name)
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

	m.start()
}
