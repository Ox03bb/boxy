package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	switch os.Args[1] {
	case "run":
		ini()
	case "prosses":
		prosses()
	default:
		panic("Unknown command")
	}

}

func ini() {
	fmt.Printf("[running] %v \n", os.Args[2])

	cmd := exec.Command("/proc/self/exe", append([]string{"prosses"}, os.Args[2:]...)...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID,
		Setctty:    true,
		Setsid:     true,
	}

	err := cmd.Run()

	erro(err)

}

func prosses() {
	fmt.Printf("[running] %v as %v\n", os.Args[2], os.Getpid())

	cmd := exec.Command(os.Args[2], os.Args[3:]...)

	syscall.Sethostname([]byte("xxx"))

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	erro(err)

}

func erro(err error) {
	if err != nil {
		panic(err)
	}
}
