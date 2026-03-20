package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/Ox03bb/boxy/internal/ipc"
	"github.com/spf13/cobra"
)

// LogsHandler builds the logs request.
func LogsHandler(cmd *cobra.Command, args []string) (*ipc.Command, error) {
	l := ipc.Logs{}

	nameflag, _ := cmd.Flags().GetString("name")
	follow, _ := cmd.Flags().GetBool("follow")
	tail, _ := cmd.Flags().GetInt("tail")

	if nameflag != "" {
		l.BoxIdentifier = nameflag
		l.Is_name = true
	} else if len(args) >= 1 {
		l.BoxIdentifier = args[0]
		l.Is_name = false
	} else {
		return nil, fmt.Errorf("box ID or name is required")
	}

	l.Follow = follow
	l.Tail = tail

	req := &ipc.Command{
		Cmd:  ipc.LogsC,
		Args: &l,
	}

	return req, nil
}

// LogsFromBox sends the request and receives a PTY fd, then streams it to stdout.
func LogsFromBox(req interface{}) error {
	sock, err := ipc.Connect("")
	if err != nil {
		return fmt.Errorf("connecting to daemon: %w", err)
	}
	defer ipc.Close(sock)

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := ipc.Send(sock, reqBytes); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	// If request asked to follow, expect a FD from daemon and stream it.
	if cmd, ok := req.(*ipc.Command); ok {
		if l, ok := cmd.Args.(*ipc.Logs); ok && l.Follow {
			unixSock, ok := sock.(*net.UnixConn)
			if !ok {
				return fmt.Errorf("not a unix socket")
			}

			fd, err := ipc.ReceiveFD(unixSock)
			if err != nil {
				return fmt.Errorf("receive fd: %w", err)
			}

			if fd <= 0 {
				return fmt.Errorf("invalid fd received")
			}

			f := os.NewFile(uintptr(fd), "pty")
			if f == nil {
				return fmt.Errorf("failed to wrap fd")
			}
			defer f.Close()

			// Stream until client interrupts or PTY closes
			_, err = io.Copy(os.Stdout, f)
			return err
		}
	}

	// Non-follow mode: read whatever the daemon writes on the socket and print it.
	resp, err := ipc.Recive(sock)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}
	if len(resp) > 0 {
		_, _ = os.Stdout.Write(resp)
	}
	return nil
}
