package server

import (
	"net"
	"os"
	"sync"
	"time"

	"github.com/Archer-01/taskmaster/internal/manager"
	"github.com/Archer-01/taskmaster/internal/utils"
)

const (
	DEL = '\r'
)

type SocketAction struct {
	socket *Socket
	add    bool
}

type Server struct {
	j       *manager.JobManager
	conns   map[*Socket]bool
	addr    string
	sock    net.Listener
	done    chan bool
	sockets chan SocketAction
}

func NewServer(addr string, m *manager.JobManager) *Server {
	var s Server

	s.addr = addr
	s.j = m
	s.done = make(chan bool, 1)
	s.sockets = make(chan SocketAction, 1)
	s.conns = make(map[*Socket]bool)

	return &s
}

func (s *Server) Stop() {
	utils.Logf("[INFO] Closing Server")
	close(s.done)
	close(s.sockets)
	s.sock.Close()
	os.Remove(s.addr)
	for client := range s.conns {
		client.Close()
	}
}

func (s *Server) Init() error {
	sock, err := net.Listen("unix", s.addr)
	if err != nil {
		return err
	}

	s.sock = sock
	return nil
}

func (s *Server) HandleSocketsList(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for {
		val, ok := <-s.sockets
		if !ok {
			return
		}

		switch val.add {
		case true:
			s.conns[val.socket] = false
		case false:
			delete(s.conns, val.socket)
		}
	}
}

func (s *Server) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	utils.Logf("[SERVER] Starting Server")

	go s.HandleSocketsList(wg)
	var er error = nil

	for {
		select {
		case <-s.done:
			return
		default:
			if er != nil {
				// this is for accept error only if server is not closed
				utils.Errorf("[SERVER] %s", er)
				time.Sleep(100 * time.Millisecond)
			}

			con, err := s.sock.Accept()
			if er = err; err != nil {
				continue
			}

			socket := NewSocket(con)

			s.sockets <- SocketAction{socket, true}

			go s.handleConnection(DEL, socket, wg)
		}
	}
}
