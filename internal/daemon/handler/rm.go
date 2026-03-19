package handler

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

	// unmount any mount points that live under the box rootfs.
	rootfs := filepath.Join(envPath, id, "rootfs")
	if err := unmountMountsUnder(rootfs); err != nil {
		// non-fatal: warn caller but continue to attempt removal
		fmt.Fprintf(sock, "Warning: some mounts could not be detached: %v\n", err)
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

// unmountMountsUnder parses /proc/self/mounts and unmounts mounts whose
// mount point path is under the provided root. Unmounts are attempted in
// reverse path length order to unmount children before parents.
func unmountMountsUnder(root string) error {
	f, err := os.Open("/proc/self/mounts")
	if err != nil {
		return err
	}
	defer f.Close()

	var mpoints []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		mpoint := fields[1]
		// normalize and compare
		if strings.HasPrefix(mpoint, root) {
			mpoints = append(mpoints, mpoint)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if len(mpoints) == 0 {
		return nil
	}

	// unmount deepest paths first
	sort.Slice(mpoints, func(i, j int) bool {
		return len(mpoints[i]) > len(mpoints[j])
	})

	var lastErr error
	for _, mp := range mpoints {
		if err := unix.Unmount(mp, unix.MNT_DETACH); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
