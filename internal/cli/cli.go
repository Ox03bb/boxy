package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Ox03bb/boxy/internal/cli/handler"
	"github.com/Ox03bb/boxy/internal/ipc"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "boxy",
	Short: "Boxy CLI",
}

func Execute() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(attachCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(execCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(psCmd)
	rootCmd.AddCommand(rmCmd)
	rootCmd.AddCommand(stopCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// ======================= Run command =======================

var runCmd = &cobra.Command{
	Use:   "run [OPTIONS] IMAGE [COMMAND]",
	Short: "Run the boxy command",
	Run: func(cmd *cobra.Command, args []string) {
		req, err := handler.RunHandler(cmd, args)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if err := handler.RunAndAttach(req); err != nil {
			fmt.Println("Error:", err)
			return
		}
	},
}

func init() {
	runCmd.Flags().String("name", "", "Assign a name to the box")
	runCmd.Flags().String("image", "", "Image to use (optional)")
}

// ======================= Attach command =======================

var attachCmd = &cobra.Command{
	Use:   "attach [OPTIONS] BOX",
	Short: "Attach to a running box",
	Run: func(cmd *cobra.Command, args []string) {
		req, err := handler.AttachHandler(cmd, args)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if err := handler.AttachToBox(req); err != nil {
			fmt.Println("Error:", err)
			return
		}
	},
}

func init() {
	attachCmd.Flags().String("name", "", "attach to a box by name instead of ID")
}

var logsCmd = &cobra.Command{
	Use:   "logs [OPTIONS] BOX",
	Short: "Show logs for a box (streams PTY output)",
	Run: func(cmd *cobra.Command, args []string) {
		req, err := handler.LogsHandler(cmd, args)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if err := handler.LogsFromBox(req); err != nil {
			fmt.Println("Error:", err)
			return
		}
	},
}

func init() {
	logsCmd.Flags().String("name", "", "use name instead of ID to identify the box")
	logsCmd.Flags().BoolP("follow", "f", false, "Follow log output")
	logsCmd.Flags().IntP("tail", "n", 0, "Show last N lines")
}

// ======================= Exec command =======================

var execCmd = &cobra.Command{
	Use:   "exec [OPTIONS] BOX COMMAND",
	Short: "Run a command in a running box (like docker exec)",
	Run: func(cmd *cobra.Command, args []string) {
		req, err := handler.ExecHandler(cmd, args)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if err := handler.RunAndAttach(req); err != nil {
			fmt.Println("Error:", err)
			return
		}
	},
}

func init() {
	execCmd.Flags().BoolP("tty", "t", false, "Allocate a pseudo-TTY")
	execCmd.Flags().BoolP("interactive", "i", false, "Keep STDIN open")
	execCmd.Flags().String("name", "", "use name instead of ID to identify the box")
}

// ======================= Ps command =======================

var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List running boxes",
	Run: func(cmd *cobra.Command, args []string) {
		if err := handler.PsHandler(cmd, args); err != nil {
			fmt.Println("Error:", err)
			return
		}
	},
}

var rmCmd = &cobra.Command{
	Use:   "rm [OPTIONS] BOX",
	Short: "Remove a box by ID or name",
	Run: func(cmd *cobra.Command, args []string) {
		if err := handler.RmHandler(cmd, args); err != nil {
			fmt.Println("Error:", err)
			return
		}
	},
}

func init() {
	rmCmd.Flags().String("name", "", "remove a box by name instead of ID")
}

// ======================= Stop command =======================

var stopCmd = &cobra.Command{
	Use:   "stop [OPTIONS] BOX",
	Short: "Stop (kill) a running box but keep rootfs and metadata",
	Run: func(cmd *cobra.Command, args []string) {
		if err := handler.StopHandler(cmd, args); err != nil {
			fmt.Println("Error:", err)
			return
		}
	},
}

func init() {
	stopCmd.Flags().String("name", "", "stop a box by name instead of ID")
}

// ======================= Start command =======================

var startCmd = &cobra.Command{
	Use:   "start [OPTIONS] BOX",
	Short: "Start a stopped/exited box",
	Run: func(cmd *cobra.Command, args []string) {
		req, err := handler.StartHandler(cmd, args)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// if attach flag provided, reuse RunAndAttach which waits for FD
		attachFlag, _ := cmd.Flags().GetBool("attach")
		if attachFlag {
			if err := handler.RunAndAttach(req); err != nil {
				fmt.Println("Error:", err)
				return
			}
			return
		}

		// otherwise just send request and print response
		sock, err := ipc.Connect("")
		if err != nil {
			fmt.Println("Error connecting to daemon:", err)
			return
		}
		defer ipc.Close(sock)

		b, _ := json.Marshal(req)
		if err := ipc.Send(sock, b); err != nil {
			fmt.Println("Error sending request:", err)
			return
		}
		resp, err := ipc.Recive(sock)
		if err == nil && len(resp) > 0 {
			fmt.Println(string(resp))
		}
	},
}

func init() {
	startCmd.Flags().BoolP("attach", "a", false, "Attach to the box after starting")
	startCmd.Flags().String("name", "", "use name instead of ID to identify the box")
}
