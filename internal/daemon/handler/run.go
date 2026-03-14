package handler

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/Ox03bb/boxy/internal/ipc"
)

func RunHandler(c ipc.Command) {
	print("========= RunHandler =========\n")

	cmnd := c.Args.(*ipc.Run).Image.Cmd
	if len(cmnd) == 0 {
		cmnd = []string{"/bin/sh"}
	}
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, cmnd...)...)

	print("1111111111111111111111111111111111111111111\n")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	print("................................\n")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWIPC,
		Unshareflags: syscall.CLONE_NEWUTS,
		// Setctty:      true,
		// Setsid:       true,
	}
	print("00000000000000000000000000000000000000000000000\n")

	err := cmd.Run()
	if err != nil {
		panic("Error: " + err.Error())
	}
}

// func ini() {
// 	fmt.Printf("[running] %v \n", os.Args[2])

// 	cmd := exec.Command("/proc/self/exe", append([]string{"prosses"}, os.Args[2:]...)...)

// 	cmd.Stdin = os.Stdin
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr

// 	cmd.SysProcAttr = &syscall.SysProcAttr{
// 		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
// 		Setctty:    true,
// 		Setsid:     true,
// 	}

// 	err := cmd.Run()

// 	if err != nil {
// 		panic("Error: " + err.Error())
// 	}
// }

// func prosses() {
// 	fmt.Printf("[running] %v as %v\n", os.Args[2], os.Getpid())

// 	cmd := exec.Command(os.Args[2], os.Args[3:]...)

// 	syscall.Sethostname([]byte("xxx"))

// 	cmd.Stdin = os.Stdin
// 	cmd.Stdout = os.Stdout
// 	cmd.Stderr = os.Stderr

// 	err := cmd.Run()

// }
