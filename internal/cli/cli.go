package cli

import (
	"fmt"
	"os"

	"github.com/Ox03bb/boxy/internal/cli/handler"
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

		if err := handler.RunAndAttach(req); err != nil {
			fmt.Println("Error:", err)
			return
		}
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
