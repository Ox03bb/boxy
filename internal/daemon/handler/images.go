package handler

import (
	"fmt"
	"net"
	"strings"
	"text/tabwriter"

	bx "github.com/Ox03bb/boxy/internal/box"
	"github.com/Ox03bb/boxy/internal/ipc"
)

// ImagesHandler writes a table of available images to the provided conn.
func ImagesHandler(c ipc.Command, sock net.Conn) {
	if sock == nil {
		return
	}

	imgs, err := bx.ListImages()
	if err != nil {
		fmt.Fprintf(sock, "Error: %v\n", err)
		return
	}

	tw := tabwriter.NewWriter(sock, 0, 8, 2, ' ', 0)
	defer tw.Flush()

	fmt.Fprintf(tw, "NAME\tCMD\n")

	for _, im := range imgs {
		cmdStr := ""
		if len(im.Cmd) > 0 {
			cmdStr = strings.Join(im.Cmd, " ")
		}
		fmt.Fprintf(tw, "%s\t%s\n", im.Name, cmdStr)
	}
}
