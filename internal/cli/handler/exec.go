package handler

import (
    "encoding/json"
    "fmt"
    "github.com/Ox03bb/boxy/internal/ipc"
    "github.com/spf13/cobra"
)

func ExecHandler(cmd *cobra.Command, args []string) (*ipc.Command, error) {
    // parse flags
    tty, _ := cmd.Flags().GetBool("tty")
    interactive, _ := cmd.Flags().GetBool("interactive")
    nameFlag, _ := cmd.Flags().GetString("name")

    if len(args) < 2 {
        return nil, fmt.Errorf("usage: exec [OPTIONS] BOX COMMAND")
    }

    boxIdentifier := args[0]
    if nameFlag != "" {
        boxIdentifier = nameFlag
    }

    command := args[1:]

    execArgs := &ipc.Exec{
        BoxIdentifier: boxIdentifier,
        Is_name:       nameFlag != "",
        Cmd:           command,
        Tty:           tty,
        Interactive:   interactive,
    }

    // build command object
    req := &ipc.Command{
        Cmd:  ipc.ExecC,
        Args: execArgs,
    }

    // ensure it's JSON-marshallable (early check)
    if _, err := json.Marshal(req); err != nil {
        return nil, fmt.Errorf("failed to marshal exec request: %w", err)
    }

    return req, nil
}
