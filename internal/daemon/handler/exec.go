package handler

import (
	"fmt"
	"net"
	"os/exec"
	"syscall"
	"time"

	bx "github.com/Ox03bb/boxy/internal/box"
	"github.com/Ox03bb/boxy/internal/ipc"
	"github.com/creack/pty"
)

func ExecHandler(c ipc.Command, sock net.Conn) {
	e, ok := c.Args.(*ipc.Exec)
	if !ok {
		fmt.Fprintf(sock, "ERR: invalid args for exec\n")
		return
	}

	boxObj, err := bx.ResolveBoxIdentifier(e.BoxIdentifier, e.Is_name)
	if err != nil {
		fmt.Fprintf(sock, "ERR: %v\n", err)
		return
	}

	args := []string{"child"}
	args = append(args, "--name")
	args = append(args, boxObj.Name)
	args = append(args, "--rootfs")
	args = append(args, boxObj.Root)
	args = append(args, "--id")
	args = append(args, boxObj.ID)
	args = append(args, e.Cmd...)

	cmd := exec.Command("/proc/self/exe", args...)

	master, slave, err := pty.Open()
	if err != nil {
		fmt.Fprintf(sock, "ERR: pty open: %v\n", err)
		return
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

	uconn, ok := sock.(*net.UnixConn)
	if !ok {
		fmt.Fprintf(sock, "ERR: not unix socket\n")
		return
	}

	if err := ipc.SendFD(uconn, int(master.Fd())); err != nil {
		fmt.Fprintf(sock, "ERR: send fd: %v\n", err)
		return
	}

	// update box metadata (store last pty path)
	boxObj.Pty = master.Name()
	_ = bx.WriteBoxJSON(boxObj)

	// start the command
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(sock, "ERR: start exec: %v\n", err)
		return
	}

	// record PID in box metadata
	if cmd.Process != nil {
		pid := cmd.Process.Pid
		boxObj.PIDs = append(boxObj.PIDs, pid)
	}
	boxObj.Status = bx.Running
	boxObj.Created_at = time.Now()
	_ = bx.WriteBoxJSON(boxObj)
}
