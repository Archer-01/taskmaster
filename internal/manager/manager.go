package manager

import (
	"fmt"
	"os"

	"github.com/Archer-01/taskmaster/internal/job"
	"github.com/Archer-01/taskmaster/internal/parser/config"
)

type JobManager struct {
	Jobs []*job.Job
}

func Init(conf config.Config) *JobManager {
	var manager JobManager
	var jobs []*job.Job

	for name, prog := range conf.Programs {
		jobs = append(jobs, job.NewJob(name, prog))
	}

	manager.Jobs = jobs
	return &manager
}

func (ins *JobManager) Start() {
	for _, j := range ins.Jobs {
		j.StartJob()
	}
}

func (ins *JobManager) Finish() {
	for _, j := range ins.Jobs {
		fmt.Fprintf(os.Stdout, "Exiting [%s]\n", j.Name)
		err := j.Command.Process.Kill()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}
	}
}
