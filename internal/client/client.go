package client

import (
	"net"

	"github.com/Archer-01/taskmaster/internal/server"
)

type Client struct {
	Socket net.Conn
}

func NewClient(socket string) (*Client, error) {
	var client Client

	con, err := net.Dial("unix", socket)
	if err != nil {
		return &client, err
	}
	client.Socket = con

	return &client, nil
}

func (s *Client) Close() {
	s.Socket.Close()
}

func (s *Client) Send(line string) error {
	_, err := s.Socket.Write([]byte(line + string(server.DEL)))
	return err
}
