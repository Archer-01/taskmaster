package client

import (
	"bufio"
	"io"
	"net"

	"github.com/Archer-01/taskmaster/internal/manager"
	"github.com/Archer-01/taskmaster/internal/server"
)

type Client struct {
	Socket net.Conn
	Rd     *bufio.Reader
	Buf    string
}

func NewClient(socket string) (*Client, error) {
	var client Client

	con, err := net.Dial("unix", socket)
	if err != nil {
		return &client, err
	}
	client.Socket = con
	client.Rd = bufio.NewReader(con)

	return &client, nil
}

func (s *Client) Close() {
	s.Socket.Close()
}

func (s *Client) Send(line string) error {
	_, err := s.Socket.Write([]byte(line + string(server.DEL)))
	return err
}

func (c *Client) Read(del byte) *manager.Response {
	for {
		line, err := c.Rd.ReadString(del)
		if err == io.EOF {
			return manager.NewResponse()
		}

		if err != nil {
			return manager.BadRequest(err)
		}

		size := len(line)
		if size == 0 {
			continue
		}

		if line[size-1] != del {
			c.Buf += line
			continue
		}

		line = c.Buf + line[:size-1]
		c.Buf = ""

		return manager.NewResponseWithBody(line)
	}
}
