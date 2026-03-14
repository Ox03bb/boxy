package cli

import (
	"fmt"

	"github.com/Ox03bb/boxy/internal/box"
	"github.com/Ox03bb/boxy/internal/ipc"
	"github.com/spf13/cobra"
)

// ParseRunArgs parses run command arguments and flags
func ParseRunArgs(cmd *cobra.Command, args []string) (*ipc.Run, error) {
	var image string
	var name string

	// Handle --image flag
	imgFlag, _ := cmd.Flags().GetString("image")
	if imgFlag != "" {
		image = imgFlag
	} else if len(args) > 0 {
		image = args[0]
	}

	// Handle --name flag
	nameFlag, _ := cmd.Flags().GetString("name")
	if nameFlag != "" {
		name = nameFlag
	} else if len(args) > 1 {
		name = args[1]
	}

	if image == "" {
		return nil, fmt.Errorf("image is required")
	}

	var command []string
	if len(args) > 2 {
		command = args[2:]
	}

	if len(command) == 0 {
		command = []string{"/bin/bash"}
	}

	run := &ipc.Run{
		Image: box.Image{Name: image, Cmd: command},
		Name:  name,
	}
	return run, nil
}
