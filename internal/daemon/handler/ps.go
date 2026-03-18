package handler

import (
	"fmt"
	"net"
	"text/tabwriter"

	bx "github.com/Ox03bb/boxy/internal/box"
	"github.com/Ox03bb/boxy/internal/ipc"
)

// PsHandler writes a docker-like table of running boxes to the provided conn.
func PsHandler(c ipc.Command, sock net.Conn) {
	if sock == nil {
		return
	}

	boxes, err := bx.LoadAllBoxes()
	if err != nil {
		fmt.Fprintf(sock, "Error: %v\n", err)
		return
	}

	// use tabwriter so columns align regardless of content widths
	tw := tabwriter.NewWriter(sock, 0, 8, 2, ' ', 0)
	defer tw.Flush()

	// Header
	fmt.Fprintf(tw, "CONTAINER ID\tNAME\tPTY\tIMAGE\n")

	for _, b := range boxes {
		id := b.ID
		if len(id) > 12 {
			id = id[:12]
		}
		image := ""
		if b.Image.Name != "" {
			image = b.Image.Name
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", id, b.Name, b.Pty, image)
	}
}
