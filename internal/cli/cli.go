package cli


import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "boxy",
	Short: "Boxy CLI",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the boxy command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running boxy...")
	},
}

func Execute() {
	rootCmd.AddCommand(runCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}