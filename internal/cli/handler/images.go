package handler

import (
	"encoding/json"
	"fmt"

	"github.com/Ox03bb/boxy/internal/ipc"
	"github.com/spf13/cobra"
)

// ImagesHandler constructs an images command, sends it to the daemon, and prints the response.
func ImagesHandler(cmd *cobra.Command, args []string) error {
	req := &ipc.Command{Cmd: ipc.ImagesC}

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

	resp, err := ipc.Recive(sock)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	fmt.Print(string(resp))
	return nil
}
