package job

import (
	"fmt"
)

const (
	STOPPED  = "STOPPED"
	STARTING = "STARTING"
	RUNNING  = "RUNNING"
	BACKOFF  = "BACKOFF"
	STOPPING = "STOPPING"
	EXITED   = "EXITED"
	FATAL    = "FATAL"
	UNKNOWN  = "UNKNOWN"
)

const (
	AUTORESTART_FALSE      = "false"
	AUTORESTART_UNEXPECTED = "unexpected"
	AUTORESTART_TRUE       = "true"
)

func (j *Job) SetState(state string) error {
	switch state {
	case STARTING, RUNNING, BACKOFF, STOPPING, EXITED, FATAL, UNKNOWN:
		j.State = state
	case STOPPED:
		j.State = STOPPED
	default:
		return fmt.Errorf("invalid state: %s", state)
	}
	return nil
}

func (j *Job) Is(state string) bool {
	return j.State == state
}
