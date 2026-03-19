package handler

import (
	"fmt"
	"net"

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

	// Prefer the in-memory runtime entry (which holds the live `*os.File` PTY).
	// Resolve from disk may not contain the PTY because it's not JSON-serialised.
	if rt != nil {
		if memBox, ok := rt.Get(boxObj.ID); ok && memBox != nil && memBox.Pty != nil {
			f := memBox.Pty
			uconn, ok := sock.(*net.UnixConn)
			if !ok {
				fmt.Fprintf(sock, "ERR: not unix socket\n")
				return
			}
			if err := ipc.SendFD(uconn, int(f.Fd())); err != nil {
				fmt.Fprintf(sock, "ERR: send fd: %v\n", err)
				return
			}
			return
		}
	}

	// Fall back to disk-resolved box.Pty (may be nil).
	if boxObj.Pty == nil {
		fmt.Fprintf(sock, "ERR: no pty\n")
		return
	}

	// use the stored PTY file handle (do not close it here; runtime/owner manages lifecycle)
	f := boxObj.Pty

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
