package handler

import (
	"encoding/json"
	"fmt"

	"github.com/Ox03bb/boxy/internal/ipc"
	"github.com/spf13/cobra"
)

// StopHandler constructs a stop command, sends it to the daemon, and prints the response.
func StopHandler(cmd *cobra.Command, args []string) error {
	s := ipc.Stop{}

	nameflag, _ := cmd.Flags().GetString("name")

	if nameflag != "" {
		s.BoxIdentifier = nameflag
		s.Is_name = true
	} else if len(args) >= 1 {
		s.BoxIdentifier = args[0]
		s.Is_name = false
	} else {
		return fmt.Errorf("box ID or name is required")
	}

	req := &ipc.Command{Cmd: ipc.StopC, Args: s}

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
