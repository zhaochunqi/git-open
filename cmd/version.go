package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is the program version
	Version = "dev"
	// CommitHash is the Git commit hash at build time
	CommitHash = "none"
	// BuildDate is the build date
	BuildDate = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show program version information",
	Long:  `Show program version, Git commit hash at build time, and build date.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "Version: %s\n", Version)
		fmt.Fprintf(cmd.OutOrStdout(), "Git Commit: %s\n", CommitHash)
		fmt.Fprintf(cmd.OutOrStdout(), "Build Date: %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
