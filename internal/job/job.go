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
	Umask   string
}

func NewJob(name string, prog *config.Program) *Job {
	cmd_list := []string{
		"sh",
		"-c",
		fmt.Sprintf("umask %v && %v", prog.Umask, prog.Command),
	}

	return &Job{
		Name:    name,
		Command: exec.Command(cmd_list[0], cmd_list[1:]...),
		Umask:   prog.Umask,
	}
}

func (j *Job) StartJob() error {
	j.Command.Stdout = os.Stdout
	j.Command.Stderr = os.Stderr

	err := j.Command.Start()
	if err != nil {
		return err
	}

	return nil
}

func (j *Job) Stop() error {
	return j.Command.Process.Kill()
}
