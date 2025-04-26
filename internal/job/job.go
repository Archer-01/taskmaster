package job

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Archer-01/taskmaster/internal/parser/config"
	"github.com/Archer-01/taskmaster/internal/utils"
)

type Job struct {
	Name          string
	Command       *exec.Cmd
	Environment   []string
	Dir           string
	Autostart     bool
	StdoutLogFile string
	StderrLogFile string
	Umask         string
}

func NewJob(name string, prog *config.Program) *Job {
	cmd_list := []string{
		"sh",
		"-c",
		fmt.Sprintf("umask %v && %v", prog.Umask, prog.Command),
	}

	return &Job{
		Name:          name,
		Command:       exec.Command(cmd_list[0], cmd_list[1:]...),
		Dir:           prog.Directory,
		Autostart:     prog.Autostart,
		Environment:   prog.Environment,
		StdoutLogFile: prog.StdoutLogFile,
		StderrLogFile: prog.StderrLogFile,
		Umask:         prog.Umask,
	}
}

func (j *Job) StartJob() error {
	if j.StdoutLogFile != "" {
		file, err := utils.OpenLogFile(j.StdoutLogFile)

		if err != nil {
			utils.Errorf(err.Error())
		}

		j.Command.Stdout = file
	} else {
		j.Command.Stdout = os.Stdout
	}

	if j.StderrLogFile != "" {
		file, err := utils.OpenLogFile(j.StderrLogFile)

		if err != nil {
			utils.Errorf(err.Error())
		}

		j.Command.Stderr = file
	} else {
		j.Command.Stderr = os.Stderr
	}

	j.Command.Env = j.Environment
	j.Command.Dir = j.Dir

	err := j.Command.Start()
	if err != nil {
		return err
	}

	return nil
}

func (j *Job) Stop() error {
	return j.Command.Process.Kill()
}
