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

	if len(os.Args) > 2 && os.Args[1] == "child" {
		child()
		return
	}

	daemon("")
}

func child() {
	cmd := &exec.Cmd{}

	name := os.Args[3]
	root := os.Args[5]

	cmd = exec.Command(os.Args[8], os.Args[9:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = []string{
		"PATH=/bin:/usr/bin:/sbin:/usr/sbin",
		"TERM=xterm-256color",
	}

	syscall.Sethostname([]byte(name))
	syscall.Chroot(root)
	syscall.Chdir("/")

	syscall.Mount("proc", "proc", "proc", 0, "")

	fmt.Printf("Running command: %s\n", cmd.String())

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
	os.Chmod(socketPath, 0660)

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

	switch cmnd.Cmd {
	case ipc.RunC:
		dh.RunHandler(cmnd, c)
	case ipc.ExecC:
		dh.ExecHandler(cmnd, c)
	case ipc.PsC:
		dh.PsHandler(cmnd, c)
	case ipc.RmC:
		dh.RmHandler(cmnd, c)
	case ipc.AttachC:
		dh.AttachHandler(cmnd, c)
	default:
		fmt.Println("Error: command not found")
	}
}
