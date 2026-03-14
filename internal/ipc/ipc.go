package ipc

import (
	"net"
)

const (
	DefaultSocketPath = "/tmp/boxyd.sock"
)

func Connect(SocketPath string) (net.Conn, error) {
	c, err := net.Dial("unix", DefaultSocketPath)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func Send(c net.Conn, msg string) error {
	_, err := c.Write([]byte(msg))
	if err != nil {
		return err
	}
	return nil
}

func Recive(c net.Conn) (int, error) {
	buf := make([]byte, 1024)
	n, err := c.Read(buf)
	if err != nil {
		return 0, err
	}
	return n, nil
}
