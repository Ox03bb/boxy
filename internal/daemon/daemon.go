package daemon

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
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

	cmd.Env = []string{
		"PATH=/bin:/usr/bin:/sbin:/usr/sbin",
	}

	syscall.Sethostname([]byte("box_01"))
	syscall.Chroot("/home/ox03bb/Desktop/boxy/env")
	syscall.Chdir("/")

	syscall.Mount("proc", "proc", "proc", 0, "")

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
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

	// handle SIGINT / SIGTERM for graceful shutdown
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigc
		fmt.Printf("received signal %v, shutting down\n", s)
		l.Close()
	}()

	for {
		cnn, err := l.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				println("listener closed, exiting")
				return nil
			}
			println("accept error", err.Error())
			return err
		}

		go handler(cnn)
	}
}

func handler(c net.Conn) {
	buf, err := ipc.Recive(c)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
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
		fmt.Println("Error: command not found")
	}
}
