package handler

import (
	"fmt"
	"net"
	"os"

	bx "github.com/Ox03bb/boxy/internal/box"
	"github.com/Ox03bb/boxy/internal/ipc"
)

func AttachHandler(c ipc.Command, sock net.Conn) {
	if sock == nil {
		return
	}

	a, ok := c.Args.(*ipc.Attach)
	if !ok {
		fmt.Fprintf(sock, "ERR: invalid args\n")
		return
	}

	boxObj, err := bx.ResolveBoxIdentifier(a.BoxIdentifier, a.Is_name)
	if err != nil {
		fmt.Fprintf(sock, "ERR: %v\n", err)
		return
	}

	if boxObj.Pty == "" {
		fmt.Fprintf(sock, "ERR: no pty\n")
		return
	}

	f, err := os.OpenFile(boxObj.Pty, os.O_RDWR, 0)
	if err != nil {
		fmt.Fprintf(sock, "ERR: open pty: %v\n", err)
		return
	}
	defer f.Close()

	uconn, ok := sock.(*net.UnixConn)
	if !ok {
		fmt.Fprintf(sock, "ERR: not unix socket\n")
		return
	}

	if err := ipc.SendFD(uconn, int(f.Fd())); err != nil {
		fmt.Fprintf(sock, "ERR: send fd: %v\n", err)
		return
	}
}
