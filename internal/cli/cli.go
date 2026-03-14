package cli

import (
	"fmt"
	"os"

	"github.com/Ox03bb/boxy/internal/ipc"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "boxy",
	Short: "Boxy CLI",
}

var runCmd = &cobra.Command{
	Use:   "run [OPTIONS] IMAGE [COMMAND]",
	Short: "Run the boxy command",
	Run: func(cmd *cobra.Command, args []string) {
		runResult, err := ParseRunArgs(cmd, args)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		req := &ipc.Command{
			Cmd:  ipc.RunC,
			Args: runResult,
		}
		if err := Client(req); err != nil {
			fmt.Println("Client error:", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	rootCmd.AddCommand(runCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
