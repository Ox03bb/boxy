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

// ================================================================

// ======================= Attach command =======================

var attachCmd = &cobra.Command{
	Use:   "attach [OPTIONS] BOX",
	Short: "Attach to a running box",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := handler.AttachHandler(cmd, args)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// if err := handler.AttachToBox(req); err != nil {
		// 	fmt.Println("Error:", err)
		// 	return
		// }
	},
}

func init() {
	attachCmd.Flags().String("name", "", "attach to a box by name instead of ID")
}

// ================================================================

func Execute() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(attachCmd)
	rootCmd.AddCommand(psCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

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
