package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version 是程序的版本号
	Version = "dev"
	// CommitHash 是构建时的 Git commit hash
	CommitHash = "none"
	// BuildDate 是构建日期
	BuildDate = "unknown"
)

// versionCmd 表示 version 命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示程序版本信息",
	Long:  `显示程序的版本号、构建时的 Git commit hash 和构建日期。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Git Commit: %s\n", CommitHash)
		fmt.Printf("Build Date: %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
