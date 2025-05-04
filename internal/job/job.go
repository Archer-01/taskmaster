package job

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/Archer-01/taskmaster/internal/parser/config"
	"github.com/Archer-01/taskmaster/internal/utils"
)

type Job struct {
	Name          string
	Command       string
	cmd           *exec.Cmd
	Environment   []string
	Dir           string
	Autostart     bool
	StdoutLogFile string
	StderrLogFile string
	Umask         string
	State         string
	StartSecs     int
	StartRetries  int
	Autorestart   string
	ExitCodes     []int
	StopSignal    syscall.Signal
	running       bool
	restarting    bool
	stopping      bool
}

func NewJob(name string, prog *config.Program) *Job {
	has_zero := false
	exit_codes := prog.ExitCodes
	for exit := range exit_codes {
		if exit == 0 {
			has_zero = true
			break
		}
	}
	if !has_zero {
		exit_codes = append(exit_codes, 0)
	}
	return &Job{
		Name:          name,
		Command:       prog.Command,
		Dir:           prog.Directory,
		Autostart:     prog.Autostart,
		Environment:   prog.Environment,
		StdoutLogFile: prog.StdoutLogFile,
		StderrLogFile: prog.StderrLogFile,
		Umask:         prog.Umask,
		State:         STOPPED,
		StartSecs:     prog.StartSecs,
		StartRetries:  prog.StartRetries,
		Autorestart:   prog.Autorestart,
		ExitCodes:     exit_codes,
		StopSignal:    utils.ParseSignal(prog.StopSignal),
		running:       false,
		stopping:      false,
	}
}

func (j *Job) Start(wg *sync.WaitGroup) {
	if j.running {
		return
	}
	go j.startJobWorker(wg)
}

func (j *Job) startJobWorker(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	j.running = true
	retries := 0
	for {
		j.SetState(STARTING)
		err := j.tryStart()
		if err != nil {
			utils.Errorf(err.Error())
			j.SetState(BACKOFF)
			retries++
			if j.StartRetries == retries {
				break
			}
			continue
		}

		cur_ts := int(time.Now().Unix())
		j.SetState(RUNNING)
		state, _ := j.cmd.Process.Wait()
		j.cmd.ProcessState = state

		if j.State == STOPPING {
			j.SetState(STOPPED)
			break
		} else if int(time.Now().Unix())-cur_ts < j.StartSecs {
			j.SetState(BACKOFF)
			retries++
			if j.StartRetries == retries {
				break
			}
			continue
		}

		j.SetState(EXITED)
		retries = 0
		if j.Autorestart == AUTORESTART_FALSE {
			break
		}
		if j.Autorestart == AUTORESTART_UNEXPECTED {
			for exit := range j.ExitCodes {
				if exit == j.cmd.ProcessState.ExitCode() {
					break
				}
			}
		}
	}
	if j.State == BACKOFF {
		j.SetState(FATAL)
	}
	j.running = false
}

func (j *Job) tryStart() error {
	cmd_list := []string{
		"sh",
		"-c",
		fmt.Sprintf("umask %v && %v", j.Umask, j.Command),
	}

	cmd := exec.Command(cmd_list[0], cmd_list[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	j.cmd = cmd

	if j.StdoutLogFile != "" {
		file, err := utils.OpenLogFile(j.StdoutLogFile)
		if err != nil {
			return err
		}

		j.cmd.Stdout = file
	} else {
		j.cmd.Stdout = os.Stdout
	}

	if j.StderrLogFile != "" {
		file, err := utils.OpenLogFile(j.StderrLogFile)
		if err != nil {
			return err
		}

		j.cmd.Stderr = file
	} else {
		j.cmd.Stderr = os.Stderr
	}

	j.cmd.Env = append(j.Environment, os.Environ()...)
	j.cmd.Dir = j.Dir

	err := j.cmd.Start()
	if err != nil {
		return err
	}

	return err
}

func (j *Job) Restart(wg *sync.WaitGroup) {
	if j.restarting {
		return
	}
	if j.running {
		go j.restartJobWorker(wg)
	}
}

func (j *Job) restartJobWorker(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	j.Stop()
	j.restarting = true
	if !j.stopping {
		j.WaitJob()
	}
	if !j.stopping {
		j.Start(wg)
	}
	j.restarting = false
}

func (j *Job) WaitJob() {
	for j.running {
		time.Sleep(100 * time.Millisecond)
	}
}

func (j *Job) Stop() {
	if j.restarting {
		j.stopping = true
	}
	if j.Is(RUNNING) {
		j.SetState(STOPPING)
		err := syscall.Kill(-j.cmd.Process.Pid, syscall.SIGINT)
		if err != nil {
			utils.Errorf(err.Error())
		}
	}
}
