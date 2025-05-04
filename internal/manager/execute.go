package manager

import "fmt"

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
	switch action {
	case QUIT, RELOAD:
		m.actions <- Action{Type: action}
		return NewResponse()
	case START, STOP, STATUS, RESTART:
		if len(args) != 1 {
			return BadRequest(fmt.Errorf("%s command must be followed by one argument", action))
		}
		j, found := m.Jobs[args[0]]
		if !found {
			return BadRequest(fmt.Errorf("%s is not recognizable", args[0]))
		}
		if action == STATUS {
			return NewResponseWithBody(j.State)
		}
		m.actions <- Action{Type: action, Args: args}
		return NewResponse()
	}
	return BadRequest(fmt.Errorf("%s Unknown command", action))
}
