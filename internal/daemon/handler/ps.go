package handler

import (
	"fmt"
	"net"
	"strings"
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
	fmt.Fprintf(tw, "ID\tIMAGE\tCOMMAND\tCREATED\tSTATUS\tPORTS\tNAME\n")

	for _, b := range boxes {
		id := b.ID
		if len(id) > 12 {
			id = id[:12]
		}

		image := ""
		if b.Image.Name != "" {
			image = b.Image.Name
		}

		cmdStr := ""
		if len(b.Image.Cmd) > 0 {
			cmdStr = strings.Join(b.Image.Cmd, " ")
		}

		created := ""
		if !b.Created_at.IsZero() {
			created = b.Created_at.Format("2006-01-02 15:04:05")
		}

		status := b.Status

		ports := ""
		if len(b.Ports) > 0 {
			var ps []string
			for k, v := range b.Ports {
				ps = append(ps, fmt.Sprintf("%s:%s", k, v))
			}
			ports = strings.Join(ps, ",")
		}

		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", id, image, cmdStr, created, status, ports, b.Name)
	}
}
