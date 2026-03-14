package daemon

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"

	"github.com/Ox03bb/boxy/internal/config"
	dh "github.com/Ox03bb/boxy/internal/daemon/handler"
	"github.com/Ox03bb/boxy/internal/ipc"
)

func StartDaemon() {
	print("StartDaemon\n")

	if len(os.Args) > 2 && os.Args[1] == "child" {
		print("chiiiiiiiiiiiiild\n")
		child()
		return
	}
	print("ddddddddddddddddddddd\n")

	daemon("")
}

func child() {
	print("\n\n")
	fmt.Printf("Args: %v\n", os.Args[2:])

	cmd := exec.Command(os.Args[2], os.Args[3:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	syscall.Sethostname([]byte("box_01"))
	syscall.Chroot("/home/ox03bb/Desktop/boxy/env")
	syscall.Chdir("/")

	// syscall.Mount("proc", "proc", "proc", 0, "")

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

	err = cmnd.UnmarshalJSON(buf)

	if err != nil {

		fmt.Printf("\033[31mError: %s\033[0m", err.Error())
		return
	}

	if cmnd.Cmd == ipc.RunC {
		dh.RunHandler(cmnd)
	} else {
		fmt.Errorf("Error: command not found")
	}
}
