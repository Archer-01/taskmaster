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
	JobManager *manager.JobManager
	Conns      map[*Socket]bool
	Addr       string
	Socket     net.Listener
	done       chan bool
	sockets    chan SocketAction
}

func NewServer(addr string, m *manager.JobManager) *Server {
	var s Server

	s.Addr = addr
	s.JobManager = m
	s.done = make(chan bool, 1)
	s.sockets = make(chan SocketAction, 1)
	s.Conns = make(map[*Socket]bool)

	return &s
}

func (s *Server) Stop() {
	utils.Logf("Closing Server")
	close(s.done)
	close(s.sockets)
	s.Socket.Close()
	os.Remove(s.Addr)
	for client := range s.Conns {
		client.Close()
	}
}

func (s *Server) Init() error {
	sock, err := net.Listen("unix", s.Addr)
	if err != nil {
		return err
	}

	s.Socket = sock
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
			s.Conns[val.socket] = false
		case false:
			delete(s.Conns, val.socket)
		}
	}
}

func (s *Server) Start(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	utils.Logf("Starting Server")

	go s.HandleSocketsList(wg)
	var er error = nil

	for {
		select {
		case <-s.done:
			return
		default:
			if er != nil {
				// this is for accept error only if server is not closed
				utils.Errorf("Server: %s", er)
				time.Sleep(100 * time.Millisecond)
			}

			con, err := s.Socket.Accept()
			if er = err; err != nil {
				continue
			}

			socket := NewSocket(con)

			s.sockets <- SocketAction{socket, true}

			go s.handleConnection(DEL, socket, wg)
		}
	}
}
