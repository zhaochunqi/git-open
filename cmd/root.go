package cmd

import (
	"fmt"
	"os"

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
	Run: func(cmd *cobra.Command, args []string) {
		// Get the Git repository in the current working directory
		repo, err := getCurrentGitDirectory()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Get the remote URL of the Git repository
		remoteURL, err := getRemoteURL(repo)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Convert the remote URL to a web URL
		webURL := convertToWebURL(remoteURL)

		// Open the web URL in the browser if the -o flag is provided
		plain, _ := cmd.Flags().GetBool("open")
		if !plain {
			err = openURLInBrowser(webURL)
			if err != nil {
				fmt.Println("Error opening URL in browser:", err)
			}
			return
		}

		// Print the web URL
		fmt.Println("Web URL:", webURL)
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
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.git-open.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolP("plain", "p", false, "Just print the web url without opening.")
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
}
