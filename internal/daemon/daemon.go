package daemon

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/Ox03bb/boxy/internal/box"
	"github.com/Ox03bb/boxy/internal/config"
	dh "github.com/Ox03bb/boxy/internal/daemon/handler"
	"github.com/Ox03bb/boxy/internal/ipc"
)

func StartDaemon() {

	if len(os.Args) > 2 && os.Args[1] == "child" {
		child()
		return
	}

	daemon("")
}

func child() {
	fmt.Printf("Args: %v\n", os.Args[2:])

	cmd := &exec.Cmd{}
	var box = box.Box{}

	cmnd := []string{}

	box.GenerateID()

	if len(os.Args) > 3 && os.Args[2] == "--name" {
		box.Name = os.Args[3]

		cmnd = append(cmnd, os.Args[4:]...)

	} else {
		box.Name = box.GenerateName(box.ID)

		cmnd = append(cmnd, os.Args[2:]...)

	}

	cmd = exec.Command(cmnd[0], cmnd[1:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = []string{
		"PATH=/bin:/usr/bin:/sbin:/usr/sbin",
		"TERM=xterm-256color",
	}

	syscall.Sethostname([]byte(box.Name))
	syscall.Chroot("/home/ox03bb/Desktop/boxy/env/e")
	syscall.Chdir("/")

	syscall.Mount("proc", "proc", "proc", 0, "")

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func daemon(socketPath string) error {
	print("Deamon\n")
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

	fmt.Printf("Received: %s\n", string(buf))

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
		dh.RunHandler(cmnd, c)
	} else {
		fmt.Println("Error: command not found")
	}
}
