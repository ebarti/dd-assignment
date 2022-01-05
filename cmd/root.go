package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const FileFlag = "file"

// Note: This file was bootstrapped using cobra init.

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dd-assignment",
	Short: "Eloi Barti's implementation of Datadog's take home assignment",
	Long:  `Eloi Barti's implementation of Datadog's take home assignment in go.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP(FileFlag, "f", "", "Path to the CSV file to process")
	rootCmd.MarkFlagRequired(FileFlag)
}
