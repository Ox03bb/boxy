package daemon

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	bx "github.com/Ox03bb/boxy/internal/box"
	"github.com/Ox03bb/boxy/internal/config"
	dh "github.com/Ox03bb/boxy/internal/daemon/handler"
	"github.com/Ox03bb/boxy/internal/ipc"
	runt "github.com/Ox03bb/boxy/internal/runtime"
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

	// initialize in-memory runtime and inject into handlers
	r := runt.New()
	dh.SetRuntime(r)

	// ensure runtime boxes are marked exited on any shutdown
	defer func() {
		ids := r.ListIDs()
		for _, id := range ids {
			if err := bx.UpdateStatus(id, bx.Exited); err != nil {
				fmt.Printf("failed to mark %s exited: %v\n", id, err)
			}
		}
	}()

	// handle SIGINT / SIGTERM for graceful shutdown
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigc
		fmt.Printf("received signal %v, shutting down\n", s)
		// mark runtime boxes exited before closing listener
		ids := r.ListIDs()
		for _, id := range ids {
			if err := bx.UpdateStatus(id, bx.Exited); err != nil {
				fmt.Printf("failed to mark %s exited: %v\n", id, err)
			}
		}
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
	defer func() {
		_ = c.Close()
	}()
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
	case ipc.StartC:
		dh.StartHandler(cmnd, c)
	case ipc.ExecC:
		dh.ExecHandler(cmnd, c)
	case ipc.StopC:
		dh.StopHandler(cmnd, c)
	case ipc.PsC:
		dh.PsHandler(cmnd, c)
	case ipc.ImagesC:
		dh.ImagesHandler(cmnd, c)
	case ipc.RmC:
		dh.RmHandler(cmnd, c)
	case ipc.AttachC:
		dh.AttachHandler(cmnd, c)
	case ipc.LogsC:
		dh.LogsHandler(cmnd, c)
	default:
		fmt.Println("Error: command not found")
	}
}
