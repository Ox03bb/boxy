package handler

import (
	"fmt"

	"github.com/Ox03bb/boxy/internal/box"
	"github.com/Ox03bb/boxy/internal/ipc"
	"github.com/spf13/cobra"
)

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
		command = []string{"/bin/sh"}
	}

	imageObj, err := box.LoadImage(image)
	if err != nil {
		return nil, fmt.Errorf("failed to load image %s: %w", image, err)
	}

	run := &ipc.Run{
		Image: *imageObj,
		Name:  name,
	}
	return run, nil
}
