package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/Ox03bb/boxy/internal/cli/handler"
	"github.com/Ox03bb/boxy/internal/ipc"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "boxy",
	Short: "Boxy CLI",
}

// run command

var runCmd = &cobra.Command{
	Use:   "run [OPTIONS] IMAGE [COMMAND]",
	Short: "Run the boxy command",
	Run: func(cmd *cobra.Command, args []string) {
		req, err := handler.RunHandler(cmd, args)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		sock, err := ipc.Connect("")
		if err != nil {
			fmt.Println("Error connecting to daemon:", err)
			return
		}
		defer ipc.Close(sock)

		// Marshal and send the command to the daemon
		reqBytes, err := json.Marshal(req)
		if err != nil {
			fmt.Println("Error marshaling request:", err)
			return
		}

		if err := ipc.Send(sock, reqBytes); err != nil {
			fmt.Println("Error sending request to daemon:", err)
			return
		}

		// Convert net.Conn to *net.UnixConn so we can receive a file descriptor.
		unixSock, ok := sock.(*net.UnixConn)
		if !ok {
			fmt.Println("Error: socket is not a unix domain socket")
			return
		}

		// Receive the PTY file descriptor from the daemon.
		fd, err := ipc.ReceiveFD(unixSock)
		if err != nil {
			fmt.Println("Error receiving FD:", err)
			return
		}

		// Create an *os.File from the received fd and attach it to a shell process.
		ptyFile := os.NewFile(uintptr(fd), "pty")
		if ptyFile == nil {
			fmt.Println("Error: failed to create file from fd")
			return
		}

		defer ptyFile.Close()

		go func() {
			io.Copy(ptyFile, os.Stdin)
		}()

		io.Copy(os.Stdout, ptyFile)
	},
}

func init() {
	// Register flags for the run command
	runCmd.Flags().String("name", "", "Assign a name to the box")
	runCmd.Flags().String("image", "", "Image to use (optional)")
}

func Execute() {
	rootCmd.AddCommand(runCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
