package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"

	"github.com/Ox03bb/boxy/internal/config"
	dh "github.com/Ox03bb/boxy/internal/daemon/handler"
	"github.com/Ox03bb/boxy/internal/ipc"
)

func StartDeamon() {

	if len(os.Args) > 2 && os.Args[0] == "child" {
		child()
		return
	}

	daemon("")
}

func child() {
	cmd := exec.Command(os.Args[2], os.Args[3:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	syscall.Sethostname([]byte("box_01"))
	syscall.Chroot("/home/ox03bb/Desktop/boxy/env")
	syscall.Chdir("/")

	syscall.Mount("proc", "proc", "proc", 0, "")

	err := cmd.Run()
	if err != nil {
		fmt.Errorf("Error: %w", err)
	}
}

func daemon(socketPath string) error {
	if socketPath == "" {
		socketPath = config.SocketPath
	}

	os.Remove(socketPath)

	l, err := net.Listen("unix", socketPath)
	if err != nil {
		println("listen error", err.Error())
		return err
	}

	defer os.Remove(socketPath)

	for {
		cnn, err := l.Accept()
		if err != nil {
			println("accept error", err.Error())
			return err
		}

		go handler(cnn)
	}
}

func handler(c net.Conn) {
	buf, err := ipc.Recive(c)

	if err != nil {
		fmt.Errorf("Error: %w", err)
		return
	}

	var cmnd ipc.Command

	err = json.Unmarshal(buf, &cmnd)

	if err != nil {
		fmt.Errorf("Error: %w", err)
		return
	}

	if cmnd.Cmd == ipc.RunC {
		dh.RunHandler(cmnd)
	} else {
		fmt.Errorf("Error: command not found")
	}
}
