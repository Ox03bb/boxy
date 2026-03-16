package handler

import (
	"net"
	"os/exec"
	"syscall"

	"github.com/Ox03bb/boxy/internal/ipc"
	"github.com/creack/pty"
)

func RunHandler(c ipc.Command, sock net.Conn) {

	cmnd := c.Args.(*ipc.Run).Image.Cmd
	if len(cmnd) == 0 {
		cmnd = []string{"/bin/sh"}
	}

	// cmd := exec.Command("/proc/self/exe", append([]string{"child"}, cmnd...)...)

	name := ""
	if len(c.Args.(*ipc.Run).Name) != 0 {
		name = c.Args.(*ipc.Run).Name
	}

	args := []string{"child"}
	if name != "" {
		args = append(args, "--name")
		args = append(args, name)
	}
	args = append(args, cmnd...)

	cmd := exec.Command("/proc/self/exe", args...)

	master, slave, err := pty.Open()
	if err != nil {
		panic("Error: " + err.Error())
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

	// sock is a net.Conn; assert to *net.UnixConn for SendFD
	if uconn, ok := sock.(*net.UnixConn); ok {
		if err := ipc.SendFD(uconn, int(master.Fd())); err != nil {
			panic("send fd error: " + err.Error())
		}
	} else {
		panic("socket is not a UnixConn")
	}

	err = cmd.Start()
	if err != nil {
		panic("Error: " + err.Error())
	}
}
