package ipc

import (
	"net"
)

const (
	DefaultSocketPath = "/tmp/boxyd.sock"
)

func Connect(SocketPath string) (net.Conn, error) {
	sock := DefaultSocketPath
	if SocketPath != "" {
		sock = SocketPath
	}

	c, err := net.Dial("unix", sock)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func Send(c net.Conn, msg []byte) error {
	_, err := c.Write(msg)
	if err != nil {
		return err
	}
	return nil
}

func Recive(c net.Conn) ([]byte, error) {
	buf := make([]byte, 2048)
	n, err := c.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[0:n], nil
}

func Close(c net.Conn) error {
	return c.Close()
}
