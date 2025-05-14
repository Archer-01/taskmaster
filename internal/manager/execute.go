package manager

import (
	"fmt"
)

const (
	STATUS = "status"
)

type Response struct {
	Data string
	Err  error
}

func NewResponse() *Response {
	return &Response{
		Data: "",
		Err:  nil,
	}
}

func NewResponseWithBody(data string) *Response {
	return &Response{
		Data: data,
		Err:  nil,
	}
}

func BadRequest(err error) *Response {
	return &Response{
		Data: "",
		Err:  err,
	}
}

func (r *Response) HasContent() bool {
	return r.Data != ""
}

func (m *JobManager) Execute(action string, args ...string) *Response {
	done := make(chan bool, 1)
	defer close(done)

	data := make(chan string, 1)
	defer close(data)

	switch action {
	case QUIT, RELOAD, START, STOP, RESTART:
		m.actions <- Action{Type: action, Done: done, Data: data, Args: args}
		success := <-done
		if success {
			return NewResponse()
		} else {
			return BadRequest(fmt.Errorf(<-data))
		}

	case STATUS:
		m.actions <- Action{Type: action, Done: done, Data: data, Args: args}
		success := <-done
		if success {
			return NewResponseWithBody(<-data)
		} else {
			return BadRequest(fmt.Errorf(<-data))
		}
	}
	return BadRequest(fmt.Errorf("%s Unknown command", action))
}
