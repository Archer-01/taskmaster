package server

import (
	"bufio"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/Archer-01/taskmaster/internal/utils"
)

type Socket struct {
	Con net.Conn
	Buf string
	Rd  *bufio.Reader
}

func NewSocket(conn net.Conn) *Socket {
	return &Socket{
		Con: conn,
		Rd:  bufio.NewReader(conn),
	}
}

func (s *Socket) Close() {
	s.Con.Close()
}

func parse(text string) (string, []string, error) {
	args := strings.FieldsFunc(text, func(r rune) bool {
		return strings.ContainsRune(" \t\n\v\f\r", r)
	})
	if len(args) < 2 {
		return args[0], make([]string, 0), nil
	}
	return args[0], args[1:], nil
}

func (_sv *Server) handleConnection(del byte, s *Socket, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	defer s.Close()

	var er error = nil

	for {
		select {
		case <-_sv.done:
			return
		default:
			if er == io.EOF {
				_sv.sockets <- SocketAction{s, false}
				return
			}

			if er != nil {
				utils.Errorf("[SERVER]: %s", er.Error())
			}

			line, err := s.Rd.ReadString(del)
			if er = err; err != nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			er = nil

			size := len(line)
			if size == 0 {
				continue
			}
			if line[size-1] != del {
				s.Buf += line
				continue
			}
			line = s.Buf + line[:size-1]
			s.Buf = ""

			cmd, args, err := parse(line)
			if err != nil {
				s.Con.Write([]byte(err.Error()))
			}

			res := _sv.j.Execute(cmd, args...)
			if res.Err != nil {
				s.Con.Write([]byte(res.Err.Error() + string(del)))
			} else {
				s.Con.Write([]byte(res.Data + string(del)))
			}
		}
	}
}
