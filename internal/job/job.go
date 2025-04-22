package job

import (
	"os"
	"os/exec"

	"github.com/Archer-01/taskmaster/internal/parser/config"
)

type Job struct {
	Name        string
	Command     *exec.Cmd
	Environment []string
}

func NewJob(name string, prog *config.Program) *Job {
	cmd_list := config.ParseCommand(prog.Command)

	return &Job{
		Name:    name,
		Command: exec.Command(cmd_list[0], cmd_list[1:]...),
		Environment: prog.Environment,
	}
}

func (j *Job) StartJob() error {
	j.Command.Stdout = os.Stdout
	j.Command.Stderr = os.Stderr
	j.Command.Env = j.Environment

	err := j.Command.Start()
	if err != nil {
		return err
	}

	return nil
}

func (j *Job) Stop() error {
	return j.Command.Process.Kill()
}
