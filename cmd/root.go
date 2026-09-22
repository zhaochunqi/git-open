package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var chdirPaths []string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "git-open",
	Short: "Open the current Git repository in your browser",
	Long: `This application retrieves the remote URL of the Git repository in the current working
directory, converts it to a web URL, and opens it in your default browser.
Pass --plain (-p) to print the URL instead of opening it.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// --version and the help command must keep working even when the config
		// file is invalid, so they bypass config validation.
		if showVersion, _ := cmd.Flags().GetBool("version"); showVersion || cmd.Name() == helpCommandName {
			return nil
		}
		if err := initConfig(); err != nil {
			return err
		}
		for _, path := range chdirPaths {
			if err := os.Chdir(path); err != nil {
				return fmt.Errorf("error changing directory to %q: %w", path, err)
			}
		}
		return nil
	},
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
	// Register the help command (with its "h" alias) before argument
	// parsing so that "git-open h" resolves.
	ensureHelpCommand(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $XDG_CONFIG_HOME/git-open/config.yaml or ~/.git-open.yaml)")
	rootCmd.PersistentFlags().StringArrayVarP(&chdirPaths, "chdir", "C", nil, "Run as if git-open was started in <path> instead of the current working directory. May be given multiple times; a non-absolute <path> is relative to the previous one.")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("plain", "p", false, "Just print the web url without opening.")
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
}

// initConfig reads the config file and the environment. It returns an error
// when the file exists but its branch configuration is invalid, so a typo
// fails the command instead of being silently ignored.
func initConfig() error {
	// Start from a clean slate so repeated calls (e.g. in tests) never pick
	// up a config file path set by an earlier run.
	viper.Reset()

	if path := configFilePath(); path != "" {
		viper.SetConfigFile(path)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	BrowserCommand = strings.TrimSpace(viper.GetString("browser"))

	hostPaths, err := parseHostBranchPaths(viper.GetStringMap("hosts"))
	if err != nil {
		HostBranchPaths = map[string]string{}
		DefaultBranchPath = ""
		return err
	}
	HostBranchPaths = hostPaths

	defaultPath, err := parseDefaultBranchPath(viper.GetString("default_style"))
	if err != nil {
		DefaultBranchPath = ""
		return err
	}
	DefaultBranchPath = defaultPath

	return nil
}

// configFilePath returns the config file to load: the --config value when
// set, otherwise the first default location that exists, or "" when there
// is no config file at all (only environment variables are used then).
func configFilePath() string {
	if cfgFile != "" {
		return cfgFile
	}
	for _, path := range defaultConfigPaths() {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	return ""
}

// defaultConfigPaths lists the default config file locations in order of
// precedence, following the XDG Base Directory specification:
// $XDG_CONFIG_HOME/git-open/config.yaml (falling back to ~/.config), then
// the legacy ~/.git-open.yaml.
func defaultConfigPaths() []string {
	var paths []string
	if dir := xdgConfigHome(); dir != "" {
		paths = append(paths, filepath.Join(dir, "git-open", "config.yaml"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".git-open.yaml"))
	}
	return paths
}

// xdgConfigHome returns $XDG_CONFIG_HOME when it is set to an absolute
// path, otherwise ~/.config. The spec requires a relative $XDG_CONFIG_HOME
// to be ignored.
func xdgConfigHome() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" && filepath.IsAbs(dir) {
		return dir
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".config")
	}
	return ""
}

func shouldAppendBranch(branchName string) bool {
	return branchName != "main" && branchName != "master"
}
