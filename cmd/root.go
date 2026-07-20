package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "git-open",
	Short: "Print the web URL of the Git repository",
	Long: `This application retrieves the remote URL of the Git repository in the current working directory
and converts it to a web URL. The web URL is then printed to the console.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		version, _ := cmd.Flags().GetBool("version")
		if version {
			fmt.Fprintf(cmd.OutOrStdout(), "Version: %s\n", Version)
			fmt.Fprintf(cmd.OutOrStdout(), "Git Commit: %s\n", CommitHash)
			fmt.Fprintf(cmd.OutOrStdout(), "Build Date: %s\n", BuildDate)
			return nil
		}
		// Get the repository, its remote URL, and the converted web URL
		repo, remoteURL, webURL, err := resolveWebURL()
		if err != nil {
			return err
		}

		branchName, err := getBranchName(repo)
		if err == nil && shouldAppendBranch(branchName) {
			// For now, we only append branch name if it's not 'main' or 'master'.
			// This can be improved later to fetch default branch from remote or allow configuration.
			webURL = buildBranchURL(webURL, branchName, remoteURL)
		}

		// Open the web URL in the browser if the -o flag is provided
		plain, _ := cmd.Flags().GetBool("plain")
		if plain {
			fmt.Fprintf(cmd.OutOrStdout(), "Web URL: %s\n", webURL)
			return nil
		}

		err = openURLInBrowserFunc(webURL)
		if err != nil {
			return fmt.Errorf("error opening URL in browser: %w", err)
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
var Execute = func() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.git-open.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("plain", "p", false, "Just print the web url without opening.")
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".git-open" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".git-open")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	BrowserCommand = strings.TrimSpace(viper.GetString("browser"))
}

func shouldAppendBranch(branchName string) bool {
	return branchName != "main" && branchName != "master"
}
