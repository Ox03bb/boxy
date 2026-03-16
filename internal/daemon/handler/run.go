package handler

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"

	"github.com/Ox03bb/boxy/internal/box"
	"github.com/Ox03bb/boxy/internal/ipc"
	"github.com/creack/pty"
)

func RunHandler(c ipc.Command, sock net.Conn) {

	var box = box.Box{}

	if c.Args == nil {
		panic("Image name is required")
	}

	box.GenerateID()

	if c.Args.(*ipc.Run).Name == "" {
		box.GenerateName()
	} else {
		box.Name = c.Args.(*ipc.Run).Name
	}
	box.SetRoot("")

	cmnd := c.Args.(*ipc.Run).Image.Cmd

	if len(cmnd) == 0 {
		cmnd = []string{"/bin/sh"}
	}

	image := c.Args.(*ipc.Run).Image

	err := image.InitFs(&box)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create rootfs:", err)
		if sock != nil {
			_ = sock.Close()
		}
		return
	}

	args := []string{"child"}

	args = append(args, "--name")
	args = append(args, box.Name)

	args = append(args, "--rootfs")
	args = append(args, box.Root)

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
