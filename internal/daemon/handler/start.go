package handler

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
	"time"

	bx "github.com/Ox03bb/boxy/internal/box"
	"github.com/Ox03bb/boxy/internal/ipc"
	"github.com/creack/pty"
)

func StartHandler(c ipc.Command, sock net.Conn) {
	if c.Args == nil {
		fmt.Fprintln(sock, "ERR: missing args for start")
		return
	}

	s, ok := c.Args.(*ipc.Start)
	if !ok {
		fmt.Fprintln(sock, "ERR: invalid args for start")
		return
	}

	boxObj, err := bx.ResolveBoxIdentifier(s.BoxIdentifier, s.Is_name)
	if err != nil {
		fmt.Fprintf(sock, "ERR: %v\n", err)
		return
	}

	if boxObj.Status == bx.Running {
		fmt.Fprintf(sock, "INFO: box %s already running\n", boxObj.ID)
		return
	}

	// choose command: prefer box.Cmd then image default then /bin/sh
	cmnd := boxObj.Cmd
	if len(cmnd) == 0 {
		cmnd = boxObj.Image.Cmd
	}
	if len(cmnd) == 0 {
		cmnd = []string{"/bin/sh"}
	}

	args := []string{"child"}
	args = append(args, "--name")
	args = append(args, boxObj.Name)
	args = append(args, "--rootfs")
	args = append(args, boxObj.Root)
	args = append(args, "--id")
	args = append(args, boxObj.ID)
	args = append(args, cmnd...)

	cmd := exec.Command("/proc/self/exe", args...)

	master, slave, err := pty.Open()
	if err != nil {
		fmt.Fprintf(sock, "ERR: pty open: %v\n", err)
		return
	}

	boxObj.Pty = master
	if rt != nil {
		_ = rt.Add(boxObj)
	}

	defer slave.Close()

	cmd.Stdin = slave
	cmd.Stdout = slave
	cmd.Stderr = slave

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC,
		Unshareflags: syscall.CLONE_NEWUTS,
		Setctty:      true,
		Setsid:       true,
		Ctty:         0,
	}

	// if attach requested, send FD over unix socket
	if s.Attach {
		uconn, ok := sock.(*net.UnixConn)
		if !ok {
			fmt.Fprintf(sock, "ERR: not unix socket\n")
			return
		}
		if err := ipc.SendFD(uconn, int(master.Fd())); err != nil {
			fmt.Fprintf(sock, "ERR: send fd: %v\n", err)
			return
		}
	}

	// write metadata before start
	if err := bx.WriteBoxJSON(boxObj); err != nil {
		fmt.Fprintln(os.Stderr, "failed to write box json:", err)
	}

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(sock, "ERR: start: %v\n", err)
		return
	}

	if cmd.Process != nil {
		pid := cmd.Process.Pid
		boxObj.PIDs = append(boxObj.PIDs, pid)
		boxObj.Status = bx.Running
		boxObj.Created_at = time.Now()
		if err := bx.WriteBoxJSON(boxObj); err != nil {
			fmt.Fprintln(os.Stderr, "failed to update box json with PID:", err)
		}
	}

	if err := bx.UpdateStatus(boxObj.ID, bx.Running); err != nil {
		fmt.Fprintln(os.Stderr, "failed to update box status to running:", err)
	}
}
