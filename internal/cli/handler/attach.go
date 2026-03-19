package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/Ox03bb/boxy/internal/ipc"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

func AttachHandler(cmd *cobra.Command, args []string) (*ipc.Command, error) {
	at := ipc.Attach{}

	nameflag, _ := cmd.Flags().GetString("name")

	if nameflag != "" {
		at.BoxIdentifier = nameflag
		at.Is_name = true
	} else if len(args) >= 1 {
		at.BoxIdentifier = args[0]
		at.Is_name = false
	} else {
		return nil, fmt.Errorf("box ID or name is required")
	}

	req := &ipc.Command{
		Cmd:  ipc.AttachC,
		Args: &at,
	}

	return req, nil
}

func AttachToBox(req interface{}) error {
	sock, err := ipc.Connect("")
	if err != nil {
		return fmt.Errorf("connecting to daemon: %w", err)
	}
	defer ipc.Close(sock)

	// Send request
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := ipc.Send(sock, reqBytes); err != nil {
		return fmt.Errorf("send: %w", err)
	}

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

	pty := os.NewFile(uintptr(fd), "pty")
	if pty == nil {
		return fmt.Errorf("failed to wrap fd")
	}
	defer pty.Close()

	// Set the remote pty window size to match local terminal so the shell
	// redraws its prompt immediately.
	if wsCols, wsRows, err := term.GetSize(int(os.Stdin.Fd())); err == nil {
		ws := &unix.Winsize{Col: uint16(wsCols), Row: uint16(wsRows)}
		_ = unix.IoctlSetWinsize(int(pty.Fd()), unix.TIOCSWINSZ, ws)
	}

	// Set terminal raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("raw mode: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Bidirectional copy
	go func() {
		_, _ = io.Copy(pty, os.Stdin)
	}()

	_, err = io.Copy(os.Stdout, pty)
	return err
}
