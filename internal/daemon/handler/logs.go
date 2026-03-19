package handler

import (
	"bytes"
	"fmt"
	"net"

	bx "github.com/Ox03bb/boxy/internal/box"
	"github.com/Ox03bb/boxy/internal/ipc"
	"golang.org/x/sys/unix"
)

// LogsHandler supports two modes:
//   - Follow: send the PTY fd to the client so it can stream live output.
//   - Tail (no follow): perform a non-blocking read of the PTY and write the
//     last N lines back over the socket.
func LogsHandler(c ipc.Command, sock net.Conn) {
	if sock == nil {
		return
	}

	l, ok := c.Args.(*ipc.Logs)
	if !ok {
		fmt.Fprintf(sock, "ERR: invalid args\n")
		return
	}

	boxObj, err := bx.ResolveBoxIdentifier(l.BoxIdentifier, l.Is_name)
	if err != nil {
		fmt.Fprintf(sock, "ERR: %v\n", err)
		return
	}

	// Prefer runtime in-memory entry which holds live PTY file handle.
	if rt != nil {
		if memBox, ok := rt.Get(boxObj.ID); ok && memBox != nil && memBox.Pty != nil {
			f := memBox.Pty

			// Follow mode: send FD to client so it can stream continuously.
			if l.Follow {
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

			// Tail mode (best-effort): perform non-blocking read of available PTY data
			fd := int(f.Fd())
			// set non-blocking
			_ = unix.SetNonblock(fd, true)
			defer unix.SetNonblock(fd, false)

			var out bytes.Buffer
			buf := make([]byte, 4096)
			for {
				n, err := unix.Read(fd, buf)
				if n > 0 {
					out.Write(buf[:n])
					continue
				}
				if err != nil {
					if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
						break
					}
					fmt.Fprintf(sock, "ERR: read: %v\n", err)
					return
				}
				// n == 0 => EOF
				break
			}

			data := out.Bytes()
			if l.Tail > 0 {
				lines := bytes.Split(data, []byte("\n"))
				if len(lines) > l.Tail {
					start := len(lines) - l.Tail
					sel := bytes.Join(lines[start:], []byte("\n"))
					// ensure trailing newline
					if len(sel) > 0 && sel[len(sel)-1] != '\n' {
						sel = append(sel, '\n')
					}
					data = sel
				}
			}

			if len(data) == 0 {
				fmt.Fprintf(sock, "")
				return
			}

			_, _ = sock.Write(data)
			return
		}
	}

	// If no runtime PTY present, we can't provide logs currently.
	fmt.Fprintf(sock, "ERR: no pty available for logs\n")
}
