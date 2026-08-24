package job

import (
	"os"
	"os/exec"
	"sync"

	"github.com/Archer-01/taskmaster/internal/logger"
	"github.com/Archer-01/taskmaster/internal/parser/config"
	"github.com/Archer-01/taskmaster/internal/utils"
)

func (j *Job) reread(prog *config.Program) (shouldRestart bool, shouldStop bool, shouldStart bool, numprocsChanged int) {
	shouldRestart = false
	shouldStop = false
	shouldStart = false
	numprocsChanged = 0

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
		shouldRestart = true
	}

	if prog.StdoutLogFile != j.StdoutLogFile {
		j.StdoutLogFile = prog.StdoutLogFile
		shouldRestart = true
	}

	if j.Autostart != prog.Autostart {
		j.Autostart = prog.Autostart
		if !j.IsRunning() && prog.Autostart {
			shouldStart = true
		} else if j.IsRunning() && !prog.Autostart {
			shouldStop = true
		}
	}
	j.ExitCodes = normalizeExitCodes(prog.ExitCodes)
	j.StopWaitSecs = prog.StopWaitSecs
	j.StopSignal = utils.ParseSignal(prog.StopSignal)
	j.Autorestart = prog.Autorestart
	j.StartSecs = prog.StartSecs
	j.StartRetries = prog.StartRetries
	j.Priority = prog.Priority
	j.ProcessName = prog.ProcessName

	if prog.RedirectStderr != j.RedirectStderr {
		j.RedirectStderr = prog.RedirectStderr
		if j.IsRunning() {
			shouldRestart = true
		}
	}

	if prog.NumProcs != j.NumProcs {
		numprocsChanged = prog.NumProcs - j.NumProcs
		j._NumProcs = prog.NumProcs
	}

	return shouldRestart, shouldStop, shouldStart, numprocsChanged
}

func (j *Job) Resize(newSize int) {
	l := len(j._running)
	if l > newSize {
		j._running = j._running[:newSize]
		j.State = j.State[:newSize]
		j.cmds = j.cmds[:newSize]
		j.pgid = j.pgid[:newSize]
		j.startReady = j.startReady[:newSize]
		j.startOnce = j.startOnce[:newSize]
	} else if l < newSize {
		cmds := make([]*exec.Cmd, newSize)
		pgid := make([]int, newSize)
		startReady := make([]chan struct{}, newSize)
		startOnce := make([]sync.Once, newSize)
		states := make([]string, newSize)

		for i := range states {
			states[i] = STOPPED
		}

		running := make([]bool, newSize)

		for i := range running {
			running[i] = false
		}

		copy(cmds, j.cmds)
		copy(pgid, j.pgid)
		copy(startReady, j.startReady)
		copy(startOnce, j.startOnce)
		copy(states, j.State)
		copy(running, j._running)

		j.cmds = cmds
		j.pgid = pgid
		j.startReady = startReady
		j.startOnce = startOnce
		j.State = states
		j._running = running
	}
}
func (j *Job) Reload(wg *sync.WaitGroup, _done chan bool, prog *config.Program) error {
	wg.Add(1)
	defer wg.Done()

	stdoutChanged := j.StdoutLogFile != prog.StdoutLogFile
	stderrChanged := j.StderrLogFile != prog.StderrLogFile
	shouldRestart, shouldStop, shouldStart, numprocsChanged := j.reread(prog)
	skipedDone := true
	if j.IsRunning() {
		if shouldStop {
			skipedDone = false
			go j.Stop(wg, _done, -1, 1)
		} else if shouldRestart {
			skipedDone = false
			go j.Restart(wg, _done, -1, 1)
		}
		if numprocsChanged > 0 {
			j.NumProcs = j._NumProcs
			j.Resize(j._NumProcs)
			go j.Start(wg, _done, j.NumProcs-numprocsChanged, numprocsChanged)
			skipedDone = false
		} else if numprocsChanged < 0 {
			num := -numprocsChanged
			ch := make(chan bool, 1)
			go j.Stop(wg, ch, j.NumProcs+numprocsChanged, num)
			go func() {
				<-ch
				logger.Debugf("Reload: resizing job to %d", j._NumProcs)
				j.Resize(j._NumProcs)
				j.NumProcs = j._NumProcs
				logger.Debug("Reload: resize done")
				_done <- true
			}()
			skipedDone = false
		}
	} else if shouldStart {
		skipedDone = false
		go j.Start(wg, _done, -1, 1)
	}

	if stdoutChanged {
		j.setLog(j.StdoutLogFile, j.StdoutWriter, os.Stdout)
	}

	if stderrChanged {
		j.setLog(j.StderrLogFile, j.StderrWriter, os.Stderr)
	}

	if skipedDone {
		_done <- true
	}
	return nil
}
