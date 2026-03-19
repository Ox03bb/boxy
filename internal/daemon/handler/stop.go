package handler

import (
	"fmt"
	"net"
	"syscall"
	"time"

	bx "github.com/Ox03bb/boxy/internal/box"
	"github.com/Ox03bb/boxy/internal/ipc"
)

// StopHandler stops (kills) the container process for a given box but preserves rootfs and box.json.
func StopHandler(c ipc.Command, sock net.Conn) {
	if sock == nil {
		return
	}

	a, ok := c.Args.(*ipc.Stop)
	if !ok {
		fmt.Fprintf(sock, "Error: invalid args for stop\n")
		return
	}

	boxObj, err := bx.ResolveBoxIdentifier(a.BoxIdentifier, a.Is_name)
	if err != nil {
		fmt.Fprintf(sock, "Error: %v\n", err)
		return
	}

	if len(boxObj.PIDs) == 0 {
		fmt.Fprintf(sock, "Error: no recorded PIDs for box\n")
		return
	}

	// attempt to stop all recorded PIDs
	for _, pid := range boxObj.PIDs {
		// try graceful
		_ = syscall.Kill(pid, syscall.SIGTERM)

		done := false
		for i := 0; i < 30; i++ {
			time.Sleep(100 * time.Millisecond)
			if err := syscall.Kill(pid, 0); err != nil {
				done = true
				break
			}
		}
		if !done {
			_ = syscall.Kill(pid, syscall.SIGKILL)
		}
	}

	// clear PIDs and update status, preserving rootfs and metadata
	boxObj.PIDs = nil
	boxObj.Status = bx.Stopped
	if err := bx.WriteBoxJSON(boxObj); err != nil {
		fmt.Fprintf(sock, "Error updating box json: %v\n", err)
		return
	}

	fmt.Fprintf(sock, "Stopped %s\n", boxObj.ID)
}
