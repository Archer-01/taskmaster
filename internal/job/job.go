package job

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/Archer-01/taskmaster/internal/logger"
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
	StopWaitSecs  int
	_running      bool
	_restarting   bool
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
		StopWaitSecs:  prog.StopWaitSecs,
		_running:      false,
		_restarting:   false,
	}
}

func (j *Job) Start(wg *sync.WaitGroup, _done chan bool) {
	defer func() { _done <- true }()
	if j.Is(STOPPING) || j._running {
		return
	}
	j._running = true
	done := make(chan bool, 1)
	defer close(done)
	go j.startJobWorker(wg, done)
	<-done
}

func (j *Job) startJobWorker(wg *sync.WaitGroup, done chan bool) {
	wg.Add(1)
	defer wg.Done()

	cmd_list := []string{
		"sh",
		"-c",
		fmt.Sprintf("umask %v && %v", j.Umask, j.Command),
	}

	cmd := exec.Command(cmd_list[0], cmd_list[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	j.cmd = cmd

	retries := 0
	for {
		j.SetState(STARTING)
		err := j.tryStart()
		if err != nil {
			logger.Error(err)
			j.SetState(BACKOFF)
			retries++
			if j.StartRetries == retries {
				break
			}
			time.Sleep(1 * time.Second)
			continue
		}

		cur_ts := int(time.Now().Unix())
		j.SetState(RUNNING)
		done <- true
		state, _ := j.cmd.Process.Wait()
		j.cmd.ProcessState = state

		if j.Is(STOPPING) {
			break
		} else if int(time.Now().Unix())-cur_ts < j.StartSecs {
			j.SetState(BACKOFF)
			retries++
			if j.StartRetries == retries {
				break
			}
			time.Sleep(1 * time.Second)
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
	if j.Is(BACKOFF) {
		j.SetState(FATAL)
	} else if j.Is(STOPPING) {
		j.SetState(STOPPED)
	}
	j._running = false
}

func (j *Job) tryStart() error {
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

	return nil
}

func (j *Job) Restart(wg *sync.WaitGroup, _done chan bool) {
	if j.Is(STOPPING) || j._restarting {
		return
	}

	j._restarting = true
	done := make(chan bool, 1)
	defer close(done)
	j.Stop(wg, done)
	j.Start(wg, _done)
	j._restarting = false
}

func (j *Job) Stop(wg *sync.WaitGroup, _done chan bool) error {
	defer func() { _done <- true }()
	if j.Is(STOPPING) {
		return nil
	}

	if j.Is(RUNNING) {
		j.SetState(STOPPING)
		cur := time.Now().Unix()
		err := syscall.Kill(-j.cmd.Process.Pid, syscall.SIGKILL)
		if err != nil {
			return err
		}

		for time.Now().Unix()-cur < int64(j.StopWaitSecs) && j._running {
			time.Sleep(100 * time.Millisecond)
		}

		if j._running {
			err = syscall.Kill(-j.cmd.Process.Pid, j.StopSignal)
			return err
		}
	}

	return nil
}
