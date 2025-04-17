package manager

import (
	"fmt"
	"os"

	"github.com/Archer-01/taskmaster/internal/job"
	"github.com/Archer-01/taskmaster/internal/parser/config"
)

type JobManager struct {
	Jobs []*job.Job
	Config *config.Config
}

func Init(conf config.Config) *JobManager {
	var manager JobManager
	var jobs []*job.Job

	for name, prog := range conf.Programs {
		jobs = append(jobs, job.NewJob(name, prog))
	}

	manager.Jobs = jobs

	manager.Config = &conf

	return &manager
}

func (ins *JobManager) Start() {
	for _, j := range ins.Jobs {
		autostart := ins.Config.Programs[j.Name].Autostart

		if autostart == false {
			continue
		}

		j.StartJob()
	}
}

func (ins *JobManager) Finish() {
	for _, j := range ins.Jobs {
		// HACK: This is a temporary fix
		// - Jobs should have a state (Started, Stopped, Running, ...etc)
		// - Status must be checked before trying to finish a job
		autostart := ins.Config.Programs[j.Name].Autostart
		if !autostart {
			continue
		}

		fmt.Fprintf(os.Stdout, "Exiting [%s]\n", j.Name)
		err := j.Command.Process.Kill()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}
	}
}
