package handler

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	bx "github.com/Ox03bb/boxy/internal/box"
	"github.com/Ox03bb/boxy/internal/config"
	"github.com/Ox03bb/boxy/internal/ipc"

	"golang.org/x/sys/unix"
)

// RmHandler removes a box identified by ID or name.
func RmHandler(c ipc.Command, sock net.Conn) {

	if sock == nil {
		return
	}

	// expect Args to be *ipc.Rm
	a, ok := c.Args.(*ipc.Rm)
	if !ok {
		fmt.Fprintf(sock, "Error: invalid args for rm\n")
		return
	}

	// resolve identifier (name, full id, or 3+ prefix)
	boxObj, err := bx.ResolveBoxIdentifier(a.BoxIdentifier, a.Is_name)
	if err != nil {
		fmt.Fprintf(sock, "Error: %v\n", err)
		return
	}
	id := boxObj.ID

	// remove the box directory under env path
	envPath := os.ExpandEnv(config.EnvPath)
	if envPath == "" {
		fmt.Fprintf(sock, "Error: EnvPath not configured\n")
		return
	}

	// try to unmount common pseudo-filesystems inside the box rootfs
	rootfs := filepath.Join(envPath, id, "rootfs")
	mounts := []string{"proc", "sys", "dev", "run", "tmp", "mnt"}
	for _, m := range mounts {
		p := filepath.Join(rootfs, m)
		if _, err := os.Stat(p); err == nil {
			// attempt lazy unmount to detach mounts that may be busy
			_ = unix.Unmount(p, unix.MNT_DETACH)
		}
	}

	boxDir := filepath.Join(envPath, id)

	// ensure the box directory exists before attempting removal
	if _, err := os.Stat(boxDir); err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(sock, "Error: box %s not found\n", id)
			return
		}
		fmt.Fprintf(sock, "Error accessing box %s: %v\n", id, err)
		return
	}

	if err := os.RemoveAll(boxDir); err != nil {
		fmt.Fprintf(sock, "Error removing box %s: %v\n", id, err)
		return
	}

	fmt.Fprintf(sock, "Removed %s\n", id)
}
