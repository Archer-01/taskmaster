package job

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Archer-01/taskmaster/internal/parser/config"
)

type Job struct {
	Name    string
	Command *exec.Cmd
}

func NewJob(name string, prog *config.Program) *Job {
	var job Job

	job.Name = name

	cmd_list := config.ParseCommand(prog.Command)

	job.Command = exec.Command(cmd_list[0], cmd_list[1:]...)

	return &job
}

func (job *Job) StartJob() {
	job.Command.Stdout = os.Stdout
	job.Command.Stderr = os.Stderr

	if err := job.Command.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return
	}
}
