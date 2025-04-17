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
	Dir     string
}

func NewJob(name string, prog *config.Program) *Job {
	cmd_list := config.ParseCommand(prog.Command)

	return &Job{
		Name: name,
		Command: exec.Command(cmd_list[0], cmd_list[1:]...),
		Dir: prog.Directory,
	}
}

func (job *Job) StartJob() {
	job.Command.Stdout = os.Stdout
	job.Command.Stderr = os.Stderr
	job.Command.Dir = job.Dir

	if err := job.Command.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return
	}
}
