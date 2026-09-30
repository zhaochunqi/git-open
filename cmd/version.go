package cmd

import (
	"fmt"
	"io"
	"strings"

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
	Long:  `Show program version, Git commit hash at build time, build date, and author.`,
	Run: func(cmd *cobra.Command, args []string) {
		printVersionInfo(cmd.OutOrStdout())
	},
}

// printVersionInfo writes the build information shared by the `version`
// subcommand and the root command's --version flag. The author line reuses
// helpAuthors, so the AUTHORS section of the help output and `git-open -v`
// never drift apart. It is omitted when no author is known.
func printVersionInfo(w io.Writer) {
	fmt.Fprintf(w, "Version: %s\n", Version)
	fmt.Fprintf(w, "Git Commit: %s\n", CommitHash)
	fmt.Fprintf(w, "Build Date: %s\n", BuildDate)
	if authors := authorList(); authors != "" {
		fmt.Fprintf(w, "Author: %s\n", authors)
	}
}

// authorList joins helpAuthors into a single line for the version output.
func authorList() string {
	nonEmpty := make([]string, 0, len(helpAuthors))
	for _, author := range helpAuthors {
		if author = strings.TrimSpace(author); author != "" {
			nonEmpty = append(nonEmpty, author)
		}
	}
	return strings.Join(nonEmpty, ", ")
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
