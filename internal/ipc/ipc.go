package ipc

import (
	"fmt"
	"io"
	"net"

	"golang.org/x/sys/unix"
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
	// If this is a unix domain socket, half-close the write side to signal EOF
	// to the peer which may be using io.ReadAll to receive the message.
	if u, ok := c.(*net.UnixConn); ok {
		if err := u.CloseWrite(); err != nil {
			return fmt.Errorf("close write error: %w", err)
		}
	}
	return nil
}

func Recive(c net.Conn) ([]byte, error) {
	return io.ReadAll(c)
}

func Close(c net.Conn) error {
	return c.Close()
}

// SendFD sends a file descriptor over a Unix domain socket using SCM_RIGHTS.
func SendFD(sock *net.UnixConn, fd int) error {
	rights := unix.UnixRights(fd)
	// send a single byte with the control message
	n, oobn, err := sock.WriteMsgUnix([]byte{0}, rights, nil)
	if err != nil {
		return err
	}
	if n != 1 || oobn != len(rights) {
		return fmt.Errorf("short write: n=%d oobn=%d want=%d", n, oobn, len(rights))
	}
	return nil
}

// ReceiveFD receives a file descriptor from a Unix domain socket.
func ReceiveFD(sock *net.UnixConn) (int, error) {
	buf := make([]byte, 1)
	oob := make([]byte, 32)
	_, oobn, _, _, err := sock.ReadMsgUnix(buf, oob)
	if err != nil {
		return -1, err
	}
	msgs, err := unix.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return -1, err
	}
	for _, msg := range msgs {
		fds, err := unix.ParseUnixRights(&msg)
		if err != nil {
			continue
		}
		if len(fds) > 0 {
			return fds[0], nil
		}
	}
	return -1, fmt.Errorf("no fd received")
}
