package handler

import (
	"fmt"

	"github.com/Ox03bb/boxy/internal/ipc"
	"github.com/spf13/cobra"
)

// StartHandler builds an ipc.Command to request starting a stopped/exited box.
func StartHandler(cmd *cobra.Command, args []string) (*ipc.Command, error) {
	st := ipc.Start{}

	nameflag, _ := cmd.Flags().GetString("name")
	if nameflag != "" {
		st.BoxIdentifier = nameflag
		st.Is_name = true
	} else if len(args) >= 1 {
		st.BoxIdentifier = args[0]
		st.Is_name = false
	} else {
		return nil, fmt.Errorf("box ID or name is required")
	}

	attach, _ := cmd.Flags().GetBool("attach")
	st.Attach = attach

	req := &ipc.Command{
		Cmd:  ipc.StartC,
		Args: &st,
	}
	return req, nil
}
