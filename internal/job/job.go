package job

import (
	"fmt"
	"io"
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
	StopWaitSecs  int
	_running      bool
	_restarting   bool
	StdoutWriter  *utils.DynamicWriter
	StderrWriter  *utils.DynamicWriter
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
		StdoutWriter:  &utils.DynamicWriter{},
		StderrWriter:  &utils.DynamicWriter{},
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
			utils.Errorf(err.Error())
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

func (j *Job) setLog(file string, writer *utils.DynamicWriter, _default io.Writer) error {
	if file != "" {
		file, err := utils.OpenLogFile(file)
		if err != nil {
			return err
		}

		writer.SetWriter(file)
	} else if _default != nil {
		writer.SetWriter(_default)
	}
	return nil
}

func (j *Job) tryStart() error {
	err := j.setLog(j.StdoutLogFile, j.StdoutWriter, os.Stdout)
	if err != nil {
		return err
	}

	err = j.setLog(j.StderrLogFile, j.StderrWriter, os.Stderr)
	if err != nil {
		return err
	}

	j.cmd.Stdout = j.StdoutWriter
	j.cmd.Stderr = j.StderrWriter

	j.cmd.Env = append(j.Environment, os.Environ()...)
	j.cmd.Dir = j.Dir

	err = j.cmd.Start()
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

func (j *Job) Reload(wg *sync.WaitGroup, _done chan bool, prog *config.Program) error {
	wg.Add(1)
	defer wg.Done()

	stdoutChanged := j.StdoutLogFile != prog.StdoutLogFile
	stderrChanged := j.StderrLogFile != prog.StderrLogFile
	shouldRestart := j.reread(prog)
	if shouldRestart && j._running {
		go j.Restart(wg, _done)
		return nil
	}

	if stdoutChanged {
		j.setLog(j.StdoutLogFile, j.StdoutWriter, os.Stdout)
	}

	if stderrChanged {
		j.setLog(j.StderrLogFile, j.StderrWriter, os.Stderr)
	}

	_done <- true
	return nil
}

func (j *Job) reread(prog *config.Program) bool {
	shouldRestart := false

	if prog.Command != j.Command {
		j.Command = prog.Command
		shouldRestart = true
	}

	if prog.Directory != j.Dir {
		j.Dir = prog.Directory
		shouldRestart = true
	}

	{
		table := make(map[string]int, len(j.Environment))
		for _, env := range j.Environment {
			table[env] += 1
		}
		for _, env := range prog.Environment {
			table[env] += 1
		}
		for _, c := range table {
			if c != 2 {
				shouldRestart = true
				j.Environment = prog.Environment
				break
			}
		}

	}

	if prog.Umask != j.Umask {
		j.Umask = prog.Umask
		shouldRestart = true
	}

	if prog.StderrLogFile != j.StderrLogFile {
		j.StderrLogFile = prog.StderrLogFile
	}

	if prog.StdoutLogFile != j.StdoutLogFile {
		j.StdoutLogFile = prog.StdoutLogFile
	}

	j.Autostart = prog.Autostart
	j.ExitCodes = prog.ExitCodes
	j.StopWaitSecs = prog.StopWaitSecs
	j.StopSignal = utils.ParseSignal(prog.StopSignal)
	j.Autorestart = prog.Autorestart
	j.StartSecs = prog.StartSecs
	j.StartRetries = prog.StartRetries

	return shouldRestart
}
