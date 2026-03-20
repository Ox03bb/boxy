package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/Ox03bb/boxy/internal/box"
	"github.com/Ox03bb/boxy/internal/ipc"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func RunAndAttach(req interface{}) error {
	sock, err := ipc.Connect("")
	if err != nil {
		return fmt.Errorf("connecting to daemon: %w", err)
	}
	defer ipc.Close(sock)

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	if err := ipc.Send(sock, reqBytes); err != nil {
		return fmt.Errorf("sending request to daemon: %w", err)
	}

	unixSock, ok := sock.(*net.UnixConn)
	if !ok {
		return fmt.Errorf("socket is not a unix domain socket")
	}

	fd, err := ipc.ReceiveFD(unixSock)
	if err != nil {
		return fmt.Errorf("receiving FD: %w", err)
	}

	ptyFile := os.NewFile(uintptr(fd), "pty")
	if ptyFile == nil {
		return fmt.Errorf("failed to create file from fd")
	}
	defer ptyFile.Close()

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("setting raw mode: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	go func() {
		_, _ = io.Copy(ptyFile, os.Stdin)
	}()

	_, err = io.Copy(os.Stdout, ptyFile)
	return err
}

func RunHandler(cmd *cobra.Command, args []string) (*ipc.Command, error) {
	runResult, err := RunArgsParse(cmd, args)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	req := &ipc.Command{
		Cmd:  ipc.RunC,
		Args: runResult,
	}

	return req, nil
}

func RunArgsParse(cmd *cobra.Command, args []string) (*ipc.Run, error) {
	var image string
	var name string

	// Handle --image flag
	imgFlag, _ := cmd.Flags().GetString("image")
	if imgFlag != "" {
		image = imgFlag
	} else if len(args) > 0 {
		image = args[0]
	}

	// Handle --name flag (name only from flag, no positional fallback)
	nameFlag, _ := cmd.Flags().GetString("name")
	if nameFlag != "" {
		name = nameFlag
	}

	if image == "" {
		return nil, fmt.Errorf("image is required")
	}

	var command []string
	if len(args) > 1 {
		command = args[1:]
	}

	if len(command) == 0 {
		command = []string{}
	}

	imageObj, err := box.LoadImage(image)
	if err != nil {
		return nil, fmt.Errorf("failed to load image %s: %w", image, err)
	}

	// Do not override the image's default here. Send CLI command in Run.Cmd
	// and let the daemon choose between Run.Cmd and Image.Cmd.
	run := &ipc.Run{
		Image: *imageObj,
		Name:  name,
		Cmd:   command,
	}
	return run, nil
}
