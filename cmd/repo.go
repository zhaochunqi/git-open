package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// repoCmd represents the repo command
var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Print the repository name",
	Long: `Print the name of the Git repository in the current working directory,
in the form of host/owner/repo (e.g. github.com/zhaochunqi/git-open).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get the repository name from the web URL of the remote
		_, _, webURL, err := resolveWebURL()
		if err != nil {
			return err
		}

		fmt.Fprintln(cmd.OutOrStdout(), repoNameFromWebURL(webURL))
		return nil
	},
}

// repoNameFromWebURL strips the scheme from a web URL,
// e.g. "https://github.com/zhaochunqi/git-open" -> "github.com/zhaochunqi/git-open".
func repoNameFromWebURL(webURL string) string {
	name := strings.TrimPrefix(webURL, "https://")
	name = strings.TrimPrefix(name, "http://")
	return strings.TrimSuffix(name, "/")
}

func init() {
	rootCmd.AddCommand(repoCmd)
}
